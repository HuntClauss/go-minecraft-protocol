package mc

import (
	"fmt"
	"io"
	"log"
	"mc-bot/mc/proto"
	"net"
	"strconv"
	"strings"
)

type ConnectionState string

const (
	ConnStateUnknown ConnectionState = "UNKNOWN"
	ConnStateStatus  ConnectionState = "STATUS"
	ConnStateLogin   ConnectionState = "LOGIN"
	ConnStatePlay    ConnectionState = "PLAY"
)

type Client struct {
	Conn              *net.TCPConn
	Version           int
	State             ConnectionState
	compressThreshold int
	Player            Player
}

func NewClient(version int) Client {
	return Client{Conn: nil, State: ConnStateUnknown, compressThreshold: -1, Version: version}
}

func (c *Client) Connect(server Server) error {
	var err error
	c.Conn, err = server.Connect()
	if err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}

	return nil
}

func (c *Client) SendPacket(pk proto.Packet) error {
	var data []byte
	if c.compressThreshold < 0 {
		data = pk.Bytes()
	} else {
		data = pk.CompressBytes(c.compressThreshold)
	}

	_, err := c.Conn.Write(data)
	return err
}

func (c *Client) RecvPacket() (proto.Packet, error) {
	if c.compressThreshold < 0 {
		return proto.NewPacketFromReader(c.Conn), nil
	}
	return proto.NewCompressPacketFromReader(c.Conn)
}

func (c *Client) Close() error {
	if c.Conn == nil {
		return fmt.Errorf("connection is nil")
	}
	return c.Conn.Close()
}

func (c *Client) Handshake(state ConnectionState) error {
	pk := proto.NewPacket(0x00)

	nextState := -1
	if state == ConnStateStatus {
		nextState = 1
	} else if state == ConnStateLogin {
		nextState = 2
	} else {
		return fmt.Errorf("invalid state. Got: %s, Required: %s or %s", state, ConnStateStatus, ConnStateLogin)
	}
	c.State = state

	addr := c.Conn.RemoteAddr().String()
	parts := strings.Split(addr, ":")
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("cannot convert port string to addres: %w", err)
	}

	err = pk.Append(
		proto.NewVarInt(c.Version),
		proto.NewString(parts[0]),
		proto.NewUShort(port),
		proto.NewVarInt(nextState),
	)
	if err != nil {
		return fmt.Errorf("cannot append data to packet object: %w", err)
	}

	if _, err = c.Conn.Write(pk.Bytes()); err != nil {
		return fmt.Errorf("cannot send handshake: %w", err)
	}
	return nil
}

func (c *Client) ServerStatus() {
	pk := proto.NewPacket(0x00)
	c.Handshake(ConnStateStatus)

	c.Conn.Write(pk.Bytes())
	c.Conn.Write([]byte{0x01, 0x00})
}

func (c *Client) Login(name, uuid string) error {
	c.Player.Name, c.Player.UUID = name, uuid
	if err := c.Handshake(ConnStateLogin); err != nil {
		return fmt.Errorf("cannot establish handshake: %w", err)
	}

	pk := proto.NewPacket(0x00)
	if err := pk.Append(proto.NewString(name)); err != nil {
		return fmt.Errorf("cannot append login start request data: %w", err)
	}

	if len(uuid) != 0 {
		if err := pk.Append(proto.NewBool(false)); err != nil {
			return fmt.Errorf("cannot append login start request data: %w", err)
		}
	} else if err := pk.Append(proto.NewBool(true), proto.NewUuidFromStr(uuid)); err != nil {
		return fmt.Errorf("cannot append login start request data: %w", err)
	}

	if _, err := c.Conn.Write(pk.Bytes()); err != nil {
		return fmt.Errorf("cannot sent login start request: %w", err)
	}
	return nil
}

func (c *Client) HandleResponses() {
	for {
		pk, err := c.RecvPacket()
		if err != nil {
			log.Println(err)
			if err == io.EOF {
				return
			}
			continue
		}

		switch c.State {
		case ConnStateStatus:
			log.Printf("TODO: State Status Response")
		case ConnStateLogin:
			err = c.handleLoginStateResponses(pk)
		case ConnStatePlay:
			err = c.handlePlayStateResponses(pk)
		default:
			err = fmt.Errorf("unknown state: %s", c.State)
		}

		if err != nil {
			log.Println(err)
		}
	}
}

func (c *Client) handleLoginStateResponses(pk proto.Packet) error {
	switch pk.ID {
	case 0x03:
		log.Printf("[INFO] Recv: %#x (set compression packet)\n", pk.ID)
		return c.HandleCompressionPacket(pk)
	case 0x02:
		log.Printf("[INFO] Recv: %#x (login success packet)\n", pk.ID)
		if err := c.HandleLoginSuccessPacket(pk); err != nil {
			return err
		}

		c.State = ConnStatePlay
		return nil
	}
	return fmt.Errorf("unknown packet in login state: %#x", pk.ID)
}

func (c *Client) handlePlayStateResponses(pk proto.Packet) error {
	switch pk.ID {

	case 0x28:
	// https://wiki.vg/Protocol#Login_(play)
	case 0x6b:
	// https://wiki.vg/Protocol#Feature_Flags
	case 0x17:
	// https://wiki.vg/Protocol#Plugin_Message
	case 0x0c:
	// https://wiki.vg/Protocol#Change_Difficulty
	case 0x34:
	// https://wiki.vg/Protocol#Player_Abilities
	case 0x4d:
	// https://wiki.vg/Protocol#Set_Held_Item
	case 0x6d:
	// https://wiki.vg/Protocol#Update_Recipes
	case 0x6e:
	// https://wiki.vg/Protocol#Update_Tags
	case 0x1c:
	// https://wiki.vg/Protocol#Entity_Event
	case 0x10:
	// https://wiki.vg/Protocol#Commands
	case 0x3d:
	// https://wiki.vg/Protocol#Update_Recipe_Book
	case 0x3c:
	// https://wiki.vg/Protocol#Synchronize_Player_Position
	case 0x45:
	// https://wiki.vg/Protocol#Server_Data
	case 0x24:
	// https://wiki.vg/Protocol#Chunk_Data_and_Update_Light
	case 0x57:
		// https://wiki.vg/Protocol#Set_Health
		c.handleSetHealthPacket(pk)
	case 0x1a:
		// https://wiki.vg/Protocol#Disconnect_(play)
	case 0x23: // keep alive
		log.Printf("[INFO] Recv: %#x (keep alive packet)", pk.ID)
		return c.HandleKeepAlivePacket(pk)
	case 0x38:
		log.Printf("[INFO] Recv: %#x (combat death packet)", pk.ID)
		return c.handleCombatDeathPacket(pk)
	case 0x62:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Sound_Effect
	case 0x42:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Set_Head_Rotation
	case 0x2c:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Update_Entity_Position_and_Rotation
	case 0x54:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Set_Entity_Velocity
	case 0x2b:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Update_Entity_Position
	case 0x27:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Update_Light
	case 0x68:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Teleport_Entity
	case 0x43:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Update_Section_Blocks
	case 0x0a:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Block_Update
	case 0x6a:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Update_Attributes
	case 0x00:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Bundle_Delimiter
	case 0x01:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Spawn_Entity
	case 0x52:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Set_Entity_Metadata
	case 0x2d:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Update_Entity_Rotation
	case 0x3e:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Remove_Entities
	case 0x5e:
		// https://wiki.vg/index.php?title=Protocol&oldid=18375#Update_Time
	case 0x56:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#Set_Experience
	case 0x69:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#Update_Advancements
	case 0x12:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#Set_Container_Content
	case 0x55:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#Set_center_Chunk
	case 0x4e:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#
	case 0x50:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#Set_Default_Spawn_Position
	case 0x22:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#Initialize_World_Border
	case 0x3a:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#Player_Info_Update
	case 0x25:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#World_Event
	case 0x18:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#Damage_Event
	case 0x41:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#Respawn
	case 0x64:
	// https://wiki.vg/index.php?title=Protocol&oldid=18375#Respawn#System_Chat_Message
	default:
		return fmt.Errorf("unknown packet id in play state: %#x", pk.ID)
	}
	return nil
}
