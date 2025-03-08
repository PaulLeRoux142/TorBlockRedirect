// Package TorBlockRedirect contains a Traefik plugin for blocking requests from the Tor network
package TorBlockRedirect

import (
	"fmt"
	"net"
)

// IPv4 is a comparable representation of a 32bit IPv4 address.
type IPv4 struct {
	addr uint32
}

// IPv6 is a comparable representation of an IPv6 address.
type IPv6 struct {
	addr [16]byte
}

// CreateIPv4 returns the IPv4 of the address a.b.c.d.
func CreateIPv4(a, b, c, d uint8) IPv4 {
	return IPv4{
		addr: uint32(a)<<24 | uint32(b)<<16 | uint32(c)<<8 | uint32(d),
	}
}

// CreateIPv6 returns the IPv6 address as a 16-byte array.
func CreateIPv6(addr [16]byte) IPv6 {
	return IPv6{
		addr: addr,
	}
}

// ParseIPv4 parses s as an IPv4 address, returning the result or an error.
func ParseIPv4(s string) (IPv4, error) {
	var fields [3]uint8
	var val, pos int
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			val = val*10 + int(s[i]) - '0'
			if val > 255 {
				return IPv4{}, fmt.Errorf("field has value >255")
			}
		} else if s[i] == '.' {
			if i == 0 || i == len(s)-1 || s[i-1] == '.' {
				return IPv4{}, fmt.Errorf("every field must have at least one digit")
			}
			if pos == 3 {
				return IPv4{}, fmt.Errorf("address too long")
			}
			fields[pos] = uint8(val)
			pos++
			val = 0
		} else {
			return IPv4{}, fmt.Errorf("unexpected character")
		}
	}
	if pos < 3 {
		return IPv4{}, fmt.Errorf("address too short")
	}
	return CreateIPv4(fields[0], fields[1], fields[2], uint8(val)), nil
}

// ParseIPv6 parses s as an IPv6 address, returning the result or an error.
func ParseIPv6(s string) (IPv6, error) {
	ip := net.ParseIP(s)
	if ip == nil || ip.To16() == nil {
		return IPv6{}, fmt.Errorf("invalid IPv6 address: %s", s)
	}

	var addr [16]byte
	copy(addr[:], ip.To16())
	return CreateIPv6(addr), nil
}

// IPv4Set contains a set of IPv4 addresses.
type IPv4Set struct {
	set map[IPv4]bool
}

// IPv6Set contains a set of IPv6 addresses.
type IPv6Set struct {
	set map[IPv6]bool
}

// CreateIPv4Set creates a new empty IPv4Set.
func CreateIPv4Set() *IPv4Set {
	return &IPv4Set{
		set: map[IPv4]bool{},
	}
}

// CreateIPv6Set creates a new empty IPv6Set.
func CreateIPv6Set() *IPv6Set {
	return &IPv6Set{
		set: map[IPv6]bool{},
	}
}

// AddIPv4 appends a new IPv4 to the set.
func (s *IPv4Set) AddIPv4(ip IPv4) {
	s.set[ip] = true
}

// AddIPv6 appends a new IPv6 to the set.
func (s *IPv6Set) AddIPv6(ip IPv6) {
	s.set[ip] = true
}

// ContainsIPv4 checks for an existing IPv4 within the set.
func (s *IPv4Set) ContainsIPv4(ip IPv4) bool {
	return s.set[ip]
}

// ContainsIPv6 checks for an existing IPv6 within the set.
func (s *IPv6Set) ContainsIPv6(ip IPv6) bool {
	return s.set[ip]
}


// Adding all IP addresses from another IPv4 set to the current set
func (s *IPv4Set) AddIPv4Set(other *IPv4Set) {
	for ip := range other.set { // Iterating over keys (IPv4 type)
		s.AddIPv4(ip) // Adding to the current set
	}
}

// Adding all IP addresses from another IPv6 set to the current set
func (s *IPv6Set) AddIPv6Set(other *IPv6Set) {
	for ip := range other.set { // Iterating over keys (IPv6 type)
		s.AddIPv6(ip) // Adding to the current set
	}
}