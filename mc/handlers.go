package mc

import (
	"fmt"
	"log"
	"mc-bot/mc/proto"
)

type ReceiveGameEvent struct {
	EventID proto.UByte
	Value   proto.Float
}

// HandleCompressionPacket https://wiki.vg/Protocol#Set_Compression
func (c *Client) HandleCompressionPacket(pk proto.Packet) error {
	threshold := proto.VarInt(-1)
	if err := pk.Scan(&threshold); err != nil {
		return fmt.Errorf("cannot scan threshold value: %w", err)
	}

	c.compressThreshold = int(threshold)
	return nil
}

func (c *Client) HandleLoginSuccessPacket(pk proto.Packet) error {
	log.Println("TODO: login success packet")
	return nil
}

func (c *Client) handleDisconnect(pk proto.Packet) error {
	var output proto.ReceivePlayerDisconnect
	if err := pk.Scan(&output); err != nil {
		return err
	}

	log.Printf("Player '%s' disconnected: %s\n", c.Player.Name, output.Reason)
	return nil
}

func (c *Client) HandleKeepAlivePacket(pk proto.Packet) error {
	var unique proto.Long
	if err := pk.Scan(&unique); err != nil {
		return err
	}

	return c.SendAlivePacket(unique)
}

func (c *Client) handleSetHealthPacket(pk proto.Packet) error {
	var health proto.SetHealthResponse
	if err := pk.Scan(&health); err != nil {
		return err
	}

	log.Printf("[INFO] Health: %v", health)
	if health.Health == 0 {
		c.PerformRespawn()
	}
	return nil
}

func (c *Client) handleCombatDeathPacket(pk proto.Packet) error {
	var death proto.CombatDeathResponse
	if err := pk.Scan(&death); err != nil {
		return err
	}

	log.Printf("[INFO] Player '%s' died. msg: %v\n", c.Player.Name, death.Message)
	c.PerformRespawn()
	return nil
}

func (c *Client) handleGameEvent(pk proto.Packet) error {
	var event ReceiveGameEvent
	if err := pk.Scan(&event); err != nil {
		return err
	}

	log.Printf("[INFO] Game Event: %v, %v", event.EventID, event.Value)
	return nil
}
