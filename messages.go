package main

type clientAuthenticate struct {
	Token string `json:"token"`
}

type clientCharacterGetId struct {
}
type clientCharacter struct {
	GetId *clientCharacterGetId `json:"get_id" esiws:""`
}

type clientLocationSubscribe struct {
}
type clientLocation struct {
	Subscribe *clientLocationSubscribe `json:"subscribe" esiws:""`
}

type clientMessage struct {
	Authenticate *clientAuthenticate `json:"authenticate" esiws:""`
	Character    *clientCharacter    `json:"character" esiws:"nested"`
	Location     *clientLocation     `json:"location" esiws:"nested"`
}

type serverError struct {
	Error string `json:"error"`
}
type serverCharacter struct {
	Id string `json:"id"`
}
type serverLocation struct {
	SolarSystemId int `json:"solar_system_id"`
}

type serverMessage struct {
	Error     *serverError     `json:"error,omitempty"`
	Character *serverCharacter `json:"character,omitempty"`
	Location  *serverLocation  `json:"location,omitempty"`
}
