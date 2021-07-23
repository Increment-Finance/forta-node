package registry

import (
	"fmt"

	"OpenZeppelin/fortify-node/clients/messaging"
	"OpenZeppelin/fortify-node/config"
	"OpenZeppelin/fortify-node/contracts"
	"OpenZeppelin/fortify-node/services/registry/regtypes"
	"OpenZeppelin/fortify-node/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
)

func (rs *RegistryService) detectAgentEvents(ethLog types.Log) error {
	log.Infof("registry agent event: tx=%s", ethLog.TxHash.Hex())
	update, agentID, ref, err := rs.detectAgentEvent(&ethLog)
	if err != nil {
		log.Errorf("registry agent error (ignoring): %s", err.Error())
		return nil
	}
	if update != nil {
		return rs.sendAgentUpdate(update, agentID, ref)
	}
	return nil
}

func (rs *RegistryService) detectAgentEvent(ethLog *types.Log) (update *agentUpdate, agentID [32]byte, ref string, err error) {

	var addedEvent *contracts.AgentRegistryAgentAdded
	addedEvent, err = rs.logUnpacker.UnpackAgentRegistryAgentAdded(ethLog)
	if err == nil {
		if (common.Hash)(addedEvent.PoolId).String() != rs.poolID.String() {
			return
		}
		return &agentUpdate{IsCreation: true}, addedEvent.AgentId, addedEvent.Ref, nil
	}

	var updatedEvent *contracts.AgentRegistryAgentUpdated
	updatedEvent, err = rs.logUnpacker.UnpackAgentRegistryAgentUpdated(ethLog)
	if err == nil {
		if (common.Hash)(updatedEvent.PoolId).String() != rs.poolID.String() {
			return
		}
		return &agentUpdate{IsUpdate: true}, updatedEvent.AgentId, updatedEvent.Ref, nil
	}

	var removedEvent *contracts.AgentRegistryAgentRemoved
	removedEvent, err = rs.logUnpacker.UnpackAgentRegistryAgentRemoved(ethLog)
	if err == nil {
		if (common.Hash)(removedEvent.PoolId).String() != rs.poolID.String() {
			return
		}
		return &agentUpdate{IsRemoval: true}, removedEvent.AgentId, "", nil
	}

	update = nil
	err = nil
	return
}

func (rs *RegistryService) sendAgentUpdate(update *agentUpdate, agentID [32]byte, ref string) error {
	agentCfg, err := rs.makeAgentConfig(agentID, ref)
	if err != nil {
		return err
	}
	update.Config = agentCfg
	log.Infof("sending agent update: %+v", update)
	rs.agentUpdates <- update
	return nil
}

func (rs *RegistryService) makeAgentConfig(agentID [32]byte, ref string) (agentCfg config.AgentConfig, err error) {
	agentCfg.ID = (common.Hash)(agentID).String()
	if len(ref) == 0 {
		return
	}

	var (
		agentData *regtypes.AgentFile
	)
	for i := 0; i < 10; i++ {
		agentData, err = rs.ipfsClient.GetAgentFile(ref)
		if err == nil {
			break
		}
	}
	if err != nil {
		err = fmt.Errorf("failed to load the agent file using ipfs ref: %v", err)
		return
	}

	var ok bool
	agentCfg.Image, ok = utils.ValidateImageRef(rs.cfg.Registry.ContainerRegistry, agentData.Manifest.ImageReference)
	if !ok {
		log.Warnf("invalid agent reference - skipping: %s", agentCfg.Image)
	}

	return
}

func (rs *RegistryService) listenToAgentUpdates() {
	for update := range rs.agentUpdates {
		rs.agentUpdatesWg.Wait()
		rs.handleAgentUpdate(update)
		rs.msgClient.Publish(messaging.SubjectAgentsVersionsLatest, rs.agentsConfigs)
	}
	close(rs.done)
}

func (rs *RegistryService) handleAgentUpdate(update *agentUpdate) {
	switch {
	case update.IsCreation:
		// Skip if we already have this agent.
		for _, agent := range rs.agentsConfigs {
			if agent.ID == update.Config.ID {
				return
			}
		}
		rs.agentsConfigs = append(rs.agentsConfigs, &update.Config)

	case update.IsUpdate:
		for _, agent := range rs.agentsConfigs {
			if agent.ID == update.Config.ID {
				agent.Image = update.Config.Image
				return
			}
		}

	case update.IsRemoval:
		var newAgents []*config.AgentConfig
		for _, agent := range rs.agentsConfigs {
			if agent.ID != update.Config.ID {
				newAgents = append(newAgents, agent)
			}
		}
		rs.agentsConfigs = newAgents

	default:
		log.Panicf("tried to handle unknown agent update")
	}
}