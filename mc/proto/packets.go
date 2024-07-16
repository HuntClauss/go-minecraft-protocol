package proto

type HandshakeRequest struct {
	ProtocolVersion VarInt
	ServerAddr      String
	ServerPort      UShort
	NextState       VarInt // NextState is an Enum - 1 for Status, 2 for Login, other values are invalid
}

type StatusResponse struct {
	Response String
}

type LoginStartRequest struct {
	Name          String // Name must be not longer than 16 characters
	HasPlayerUUID Bool
	//PlayerUUID    Uuid // PlayerUUID is optional so does not read it
}

type SetHealthResponse struct {
	Health     Float
	Food       VarInt
	Saturation Float
}

type CombatDeathResponse struct {
	PlayerID VarInt
	Message  String // Chat
	// Message  Chat // Currently I don't know how to handle this type
}

type ClientCommandActionEnum VarInt

const (
	PerformRespawn = VarInt(0)
	RequestStats   = VarInt(1)
)

type ClientCommandRequest struct {
	ActionID ClientCommandActionEnum
}
