package mc

import (
	"log"
	"mc-bot/mc/proto"
)

type RequestRespawn struct {
	TypeID proto.VarInt
}

type RequestAlivePacket struct {
	ID proto.Long
}

func (c *Client) PerformRespawn() error {
	packet := proto.NewPacket(0x07)
	if err := packet.Append(&RequestRespawn{TypeID: proto.VarInt(0)}); err != nil {
		return err
	}

	if err := c.SendPacket(packet); err != nil {
		return err
	}
	log.Printf("[INFO] Player '%s' performed respawn\n", c.Player.Name)
	return nil
}

func (c *Client) SendAlivePacket(unique proto.Long) error {
	packet := proto.NewPacket(0x12)
	if err := packet.Append(&RequestAlivePacket{ID: unique}); err != nil {
		return err
	}

	if err := c.SendPacket(packet); err != nil {
		return err
	}

	return nil
}
