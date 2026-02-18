package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// Crc16xmodem calculates CRC16-XMODEM checksum
func Crc16xmodem(data string) uint16 {
	var crc uint16 = 0xFFFF
	for i := 0; i < len(data); i++ {
		crc ^= uint16(data[i]) << 8
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc = crc << 1
			}
		}
	}
	return crc
}

func FormatTLV(tag, value string) string {
	return fmt.Sprintf("%s%02d%s", tag, len(value), value)
}

func SanitizePromptPayID(id string) string {
	reg := regexp.MustCompile(`[^0-9]`)
	return reg.ReplaceAllString(id, "")
}

func GeneratePromptPayPayload(promptpayID string, amount float64) string {
	target := SanitizePromptPayID(promptpayID)
	
	var formattedTarget string
	var targetType string
	if len(target) < 13 {
		phone := target
		if strings.HasPrefix(phone, "0") {
			phone = "66" + phone[1:]
		}
		formattedTarget = fmt.Sprintf("%013s", phone)
		targetType = "01"
	} else if len(target) == 13 {
		formattedTarget = target
		targetType = "02"
	} else {
		formattedTarget = target
		targetType = "03"
	}

	var fields []string
	fields = append(fields, FormatTLV("00", "01")) 
	fields = append(fields, FormatTLV("01", "12")) 
	
	merchantInfo := FormatTLV("00", "A000000677010111") + FormatTLV(targetType, formattedTarget)
	fields = append(fields, FormatTLV("29", merchantInfo))
	
	fields = append(fields, FormatTLV("53", "764")) 
	fields = append(fields, FormatTLV("54", fmt.Sprintf("%.2f", amount)))
	fields = append(fields, FormatTLV("58", "TH")) 
	
	payload := strings.Join(fields, "") + "6304"
	crc := Crc16xmodem(payload)
	return payload + fmt.Sprintf("%04X", crc)
}
