package net

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)	

type IPv4 struct {
	IP   uint32
	MaskLen uint32
}


func IP(ipv4str string) (ip IPv4, err error) {
	/*
		Format String IPv4 address to Uint type value.
		Input string like "172.28.0.1" or "172.28.0.1/32"
	*/

	var ipv4_prefix string
	var mask_len uint32

	if strings.Contains(ipv4str, "/") {

		ipv4_with_mask_len := strings.Split(ipv4str, "/")

		if len(ipv4_with_mask_len) > 2 {
			err = errors.New("Wrong format IP string with mask_len!")
			return
		}

		ipv4_prefix = ipv4_with_mask_len[0]
		_mask_len, _ := strconv.Atoi(ipv4_with_mask_len[1])
		mask_len = uint32(_mask_len)

	} else {
		ipv4_prefix = ipv4str
		/*Define the mask_len length out of 0-32 when not been specified.*/
		mask_len = 33
	}

	ips := strings.Split(ipv4_prefix, ".")

	if len(ips) != 4 {
		err = errors.New("Wrong format IP string!")
		return
	}

	ip1, _ := strconv.Atoi(ips[0])
	ip2, _ := strconv.Atoi(ips[1])
	ip3, _ := strconv.Atoi(ips[2])
	ip4, _ := strconv.Atoi(ips[3])

	if ip1 > 255 || ip1 < 0 || ip2 > 255 || ip2 < 0 || ip3 > 255 || ip3 < 0 || ip4 > 255 || ip4 < 0 {
		err = errors.New("IPv4 value out of range!")
		return
	}

	var ip_t uint32

	ip_t += uint32(ip1 * 0x1000000)
	ip_t += uint32(ip2 * 0x10000)
	ip_t += uint32(ip3 * 0x100)
	ip_t += uint32(ip4)

	ip.IP = ip_t
	ip.MaskLen = mask_len

	return ip, nil
}

func (ip IPv4) IPStr() string {
	/*
		Convert IPv4 Type address to string type address.
		When IPv4.MaskLen not eq 33, then output ip with mask lenght.
	*/

	ip_val := ip.IP

	if ip.MaskLen != 33 {
		return fmt.Sprintf("%d.%d.%d.%d/%d", ip_val>>24, ip_val<<8>>24, ip_val<<16>>24, ip_val<<24>>24, ip.MaskLen)
	}

	return fmt.Sprintf("%d.%d.%d.%d", ip_val>>24, ip_val<<8>>24, ip_val<<16>>24, ip_val<<24>>24)
}

func power2(n uint32) uint32{
	/*
		Calculate the value of 2**n.
		Do not compute n > 32
	*/

	if n == 0 {
		return 1
	} else {
		return 2 * power2(n-1)
	}

}


func (ip IPv4) NetworkAddr() (net IPv4){
	/*
		Return the network address of the IPv4 address;
		return a new IPv4 struct.
	*/
	
	var mask_len uint32

	if ip.MaskLen == 33{
		mask_len = 32
	} else{
		mask_len = ip.MaskLen
	}

	net.IP = ip.IP & (4294967295 - power2(32-mask_len) + 1)
	net.MaskLen = mask_len

	return
}

func main() {
/*	t, err := StrtoIPv4("172.255.20.1")
	_ = err
	fmt.Println(IPv4toStr(t))
	fmt.Println(t.NetworkAddr())*/
	var val uint32

	val =  16777729 & 4294967040
	t := IPv4{}
	t.IP = val
	t.MaskLen = 32
	fmt.Println(t.IPStr(), t.NetworkAddr())
}
