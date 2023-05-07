package discord

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"go.uber.org/zap"
)

type VoiceManager struct {
	log *zap.Logger
}

func NewVoiceManager(log *zap.Logger) *VoiceManager {
	return &VoiceManager{
		log: log,
	}
}

func (vm *VoiceManager) Log() *zap.Logger {
	return vm.log
}

// ConvertMp3ToBuffer plays the current buffer to the provided channel.
func (vm *VoiceManager) ConvertMp3ToBuffer(filePath string, buffer *[][]byte) error {
	// loadSound attempts to load an encoded sound file from disk.

	vm.Log().Info("Converting Mp3 to buffer")
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening mp3 file :", err)
		return err
	}

	// Convert from MP3 to DCA
	vm.Log().Info("DCA Encode")
	encodeSession, err := dca.EncodeFile(filePath, dca.StdEncodeOptions)
	if err != nil {
		fmt.Println("Error opening encoding session file :", err)
		return err
	}

	// Make sure everything is cleaned up, that for example the encoding process if any issues happened isnt lingering around
	defer encodeSession.Cleanup()

	// output, err := os.Create("output.dca")
	// if err != nil {
	// 	// Handle the error
	// 	vm.Log().Error("Failed to create DCA file")
	// 	return err
	// }

	// io.Copy(output, encodeSession)

	// Seek to the beginning of file
	// _, err = output.Seek(0, os.SEEK_SET)
	// if err != nil {
	// 	return err
	// }

	// decoder := dca.NewDecoder(output)

	for {
		// Read opus frame from dca file.
		vm.Log().Info("Decoding frame")
		frame, err := encodeSession.OpusFrame()
		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			vm.Log().Info("Closing file, EOF")
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		// Append encoded pcm data to the buffer.
		*buffer = append(*buffer, frame)
	}
}

func (vm *VoiceManager) PlayAudio(s *discordgo.Session, guildID string, channelID string, buffer *[][]byte) (err error) {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range *buffer {
		vm.Log().Info("sending to Opus channel", zap.ByteString("buff", buff))
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}
