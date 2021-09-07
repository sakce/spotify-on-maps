#!make
include .env
export
current_dir=$(realpath .)
export includedir = $(current_dir)/

rest:
	go run main.go

player:
	go run player/player.go 

song_name:
	go run song/song.go

song_artist:
	go run artist/song_artist.go