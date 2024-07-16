package mc

import (
	"fmt"
	"log"
	"mc-bot/mc/proto"
)

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

func (c *Client) HandleKeepAlivePacket(pk proto.Packet) error {
	var unique proto.Long
	if err := pk.Scan(&unique); err != nil {
		return err
	}

	packet := proto.NewPacket(0x12)
	if err := packet.Append(&unique); err != nil {
		return err
	}

	if err := c.SendPacket(packet); err != nil {
		return err
	}
	return nil
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

	log.Printf("[INFO]: Player '%s' died. msg: %v\n", c.Player.Name, death.Message)
	c.PerformRespawn()
	return nil
}

func (c *Client) PerformRespawn() error {
	packet := proto.NewPacket(0x07)
	if err := packet.Append(proto.NewVarInt(0)); err != nil {
		return err
	}

	if err := c.SendPacket(packet); err != nil {
		return err
	}
	log.Printf("[INFO]: Player '%s' performed respawn\n", c.Player.Name)
	return nil
}
