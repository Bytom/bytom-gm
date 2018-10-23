package common

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/bytom-gm/common/bech32"
	"github.com/bytom-gm/consensus"
)

func TestAddresses(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		encoded string
		valid   bool
		result  Address
		f       func() (Address, error)
		net     *consensus.Params
	}{
		{
			name:    "segwit testnet p2wpkh v0",
			addr:    "gm1qw508d6qejxtdg4y5r3zarvary0c5xw7kz52vyz",
			encoded: "gm1qw508d6qejxtdg4y5r3zarvary0c5xw7kz52vyz",
			valid:   true,
			result: tstAddressWitnessPubKeyHash(
				0,
				[20]byte{
					0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4, 0x54, 0x94,
					0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23, 0xf1, 0x43, 0x3b, 0xd6},
				consensus.TestNetParams.Bech32HRPSegwit),
			f: func() (Address, error) {
				pkHash := []byte{
					0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4, 0x54, 0x94,
					0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23, 0xf1, 0x43, 0x3b, 0xd6}
				return NewAddressWitnessPubKeyHash(pkHash, &consensus.TestNetParams)
			},
			net: &consensus.TestNetParams,
		},
		{
			name:    "segwit testnet p2wsh v0",
			addr:    "gm1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qn4fdhq",
			encoded: "gm1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qn4fdhq",
			valid:   true,
			result: tstAddressWitnessScriptHash(
				0,
				[32]byte{
					0x18, 0x63, 0x14, 0x3c, 0x14, 0xc5, 0x16, 0x68,
					0x04, 0xbd, 0x19, 0x20, 0x33, 0x56, 0xda, 0x13,
					0x6c, 0x98, 0x56, 0x78, 0xcd, 0x4d, 0x27, 0xa1,
					0xb8, 0xc6, 0x32, 0x96, 0x04, 0x90, 0x32, 0x62},
				consensus.TestNetParams.Bech32HRPSegwit),
			f: func() (Address, error) {
				scriptHash := []byte{
					0x18, 0x63, 0x14, 0x3c, 0x14, 0xc5, 0x16, 0x68,
					0x04, 0xbd, 0x19, 0x20, 0x33, 0x56, 0xda, 0x13,
					0x6c, 0x98, 0x56, 0x78, 0xcd, 0x4d, 0x27, 0xa1,
					0xb8, 0xc6, 0x32, 0x96, 0x04, 0x90, 0x32, 0x62}
				return NewAddressWitnessScriptHash(scriptHash, &consensus.TestNetParams)
			},
			net: &consensus.TestNetParams,
		},
		{
			name:    "segwit testnet p2wsh witness v0",
			addr:    "gm1qqqqqp399et2xygdj5xreqhjjvcmzhxw4aywxecjdzew6hylgvseslrprk6",
			encoded: "gm1qqqqqp399et2xygdj5xreqhjjvcmzhxw4aywxecjdzew6hylgvseslrprk6",
			valid:   true,
			result: tstAddressWitnessScriptHash(
				0,
				[32]byte{
					0x00, 0x00, 0x00, 0xc4, 0xa5, 0xca, 0xd4, 0x62,
					0x21, 0xb2, 0xa1, 0x87, 0x90, 0x5e, 0x52, 0x66,
					0x36, 0x2b, 0x99, 0xd5, 0xe9, 0x1c, 0x6c, 0xe2,
					0x4d, 0x16, 0x5d, 0xab, 0x93, 0xe8, 0x64, 0x33},
				consensus.TestNetParams.Bech32HRPSegwit),
			f: func() (Address, error) {
				scriptHash := []byte{
					0x00, 0x00, 0x00, 0xc4, 0xa5, 0xca, 0xd4, 0x62,
					0x21, 0xb2, 0xa1, 0x87, 0x90, 0x5e, 0x52, 0x66,
					0x36, 0x2b, 0x99, 0xd5, 0xe9, 0x1c, 0x6c, 0xe2,
					0x4d, 0x16, 0x5d, 0xab, 0x93, 0xe8, 0x64, 0x33}
				return NewAddressWitnessScriptHash(scriptHash, &consensus.TestNetParams)
			},
			net: &consensus.TestNetParams,
		},
		// Invalid segwit addresses
		{
			name:  "segwit invalid hrp",
			addr:  "tc1qw508d6qejxtdg4y5r3zarvary0c5xw7kg3g4ty",
			valid: false,
			net:   &consensus.TestNetParams,
		},
		{
			name:  "segwit mixed case",
			addr:  "tm1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sL5k7",
			valid: false,
			net:   &consensus.TestNetParams,
		},
		{
			name:  "segwit zero padding of more than 4 bits",
			addr:  "tm1pw508d6qejxtdg4y5r3zarqfsj6c3",
			valid: false,
			net:   &consensus.TestNetParams,
		},
		{
			name:  "segwit non-zero padding in 8-to-5 conversion",
			addr:  "tm1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3pjxtptv",
			valid: false,
			net:   &consensus.TestNetParams,
		},
	}

	for i, test := range tests {
		// Decode addr and compare error against valid.
		decoded, err := DecodeAddress(test.addr, test.net)
		if (err == nil) != test.valid {
			t.Errorf("example: %d :%v: decoding test failed: %v", i, test.name, err)
			return
		}

		if err == nil {
			// Ensure the stringer returns the same address as the
			// original.

			if decodedStringer, ok := decoded.(fmt.Stringer); ok {
				addr := test.addr

				// For Segwit addresses the string representation
				// will always be lower case, so in that case we
				// convert the original to lower case first.
				if strings.Contains(test.name, "segwit") {
					addr = strings.ToLower(addr)
				}

				if addr != decodedStringer.String() {
					t.Errorf("%v: String on decoded value does not match expected value: %v != %v",
						test.name, test.addr, decodedStringer.String())
					return
				}

			}

			// Encode again and compare against the original.
			encoded := decoded.EncodeAddress()
			if test.encoded != encoded {
				t.Errorf("%v: decoding and encoding produced different addressess: %v != %v",
					test.name, test.encoded, encoded)
				return
			}

			// Perform type-specific calculations.
			var saddr []byte
			switch decoded.(type) {

			case *AddressWitnessPubKeyHash:
				saddr = tstAddressSegwitSAddr(encoded)
			case *AddressWitnessScriptHash:
				saddr = tstAddressSegwitSAddr(encoded)
			}

			// Check script address, as well as the Hash160 method for P2PKH and
			// P2SH addresses.
			if !bytes.Equal(saddr, decoded.ScriptAddress()) {
				t.Errorf("%v: script addresses do not match:\n%x != \n%x",
					test.name, saddr, decoded.ScriptAddress())
				return
			}
			switch a := decoded.(type) {

			case *AddressWitnessPubKeyHash:
				if hrp := a.Hrp(); test.net.Bech32HRPSegwit != hrp {
					t.Errorf("%v: hrps do not match:\n%x != \n%x",
						test.name, test.net.Bech32HRPSegwit, hrp)
					return
				}

				expVer := test.result.(*AddressWitnessPubKeyHash).WitnessVersion()
				if v := a.WitnessVersion(); v != expVer {
					t.Errorf("%v: witness versions do not match:\n%x != \n%x",
						test.name, expVer, v)
					return
				}

				if p := a.WitnessProgram(); !bytes.Equal(saddr, p) {
					t.Errorf("%v: witness programs do not match:\n%x != \n%x",
						test.name, saddr, p)
					return
				}

			case *AddressWitnessScriptHash:
				if hrp := a.Hrp(); test.net.Bech32HRPSegwit != hrp {
					t.Errorf("%v: hrps do not match:\n%x != \n%x",
						test.name, test.net.Bech32HRPSegwit, hrp)
					return
				}

				expVer := test.result.(*AddressWitnessScriptHash).WitnessVersion()
				if v := a.WitnessVersion(); v != expVer {
					t.Errorf("%v: witness versions do not match:\n%x != \n%x",
						test.name, expVer, v)
					return
				}

				if p := a.WitnessProgram(); !bytes.Equal(saddr, p) {
					t.Errorf("%v: witness programs do not match:\n%x != \n%x",
						test.name, saddr, p)
					return
				}
			}

			// Ensure the address is for the expected network.
			if !decoded.IsForNet(test.net) {
				t.Errorf("%v: calculated network does not match expected",
					test.name)
				return
			}
		}

		if !test.valid {
			// If address is invalid, but a creation function exists,
			// verify that it returns a nil addr and non-nil error.
			if test.f != nil {
				_, err := test.f()
				if err == nil {
					t.Errorf("%v: address is invalid but creating new address succeeded",
						test.name)
					return
				}
			}
			continue
		}

		// Valid test, compare address created with f against expected result.
		addr, err := test.f()
		if err != nil {
			t.Errorf("%v: address is valid but creating new address failed with error %v",
				test.name, err)
			return
		}

		if !reflect.DeepEqual(addr, test.result) {
			t.Errorf("%v: created address does not match expected result",
				test.name)
			return
		}
	}
}

// TstAddressWitnessPubKeyHash creates an AddressWitnessPubKeyHash, initiating
// the fields as given.
func tstAddressWitnessPubKeyHash(version byte, program [20]byte,
	hrp string) *AddressWitnessPubKeyHash {

	return &AddressWitnessPubKeyHash{
		hrp:            hrp,
		witnessVersion: version,
		witnessProgram: program,
	}
}

// TstAddressWitnessScriptHash creates an AddressWitnessScriptHash, initiating
// the fields as given.
func tstAddressWitnessScriptHash(version byte, program [32]byte,
	hrp string) *AddressWitnessScriptHash {

	return &AddressWitnessScriptHash{
		hrp:            hrp,
		witnessVersion: version,
		witnessProgram: program,
	}
}

// TstAddressSegwitSAddr returns the expected witness program bytes for
// bech32 encoded P2WPKH and P2WSH bitcoin addresses.
func tstAddressSegwitSAddr(addr string) []byte {
	_, data, err := bech32.Bech32Decode(addr)
	if err != nil {
		return []byte{}
	}

	// First byte is version, rest is base 32 encoded data.
	data, err = bech32.ConvertBits(data[1:], 5, 8, false)
	if err != nil {
		return []byte{}
	}
	return data
}
