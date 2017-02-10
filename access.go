package main

type Access int

const (
	ReadOnly Access = iota
	ReadWrite
	Admin
)
