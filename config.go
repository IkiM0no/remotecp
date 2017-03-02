package main

type Config struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	RemotePath string `json:"remote_path"`
}
