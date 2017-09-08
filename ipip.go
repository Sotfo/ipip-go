package ipip

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
)

type Datx struct {
	data   []byte
	index  []byte
	flag   []uint32
	offset uint32
}

type IPIP struct {
	CR string // Country or Region              // 国家
	RG string // Region                         // 省会或直辖市（国内）
	CT string // City                           // 地区或城市 （国内）
	CY string // County                         // 学校或单位 （国内）
	IS string // Isp                            // 运营商字段（只有购买了带有运营商版本的数据库才会有）
	LA string // Latitude                       // 纬度     （每日版本提供）
	LN string // Longitude                      // 经度     （每日版本提供）
	T1 string // Timezone one                   // 时区一, 可能不存在  （每日版本提供）
	T2 string // Timezone two                   // 时区二, 可能不存在  （每日版本提供）
	AC string // Administration division code   // 中国行政区划代码    （每日版本提供）
	PC string // International phone code       // 国际电话代码        （每日版本提供）
	CC string // Country code                   // 国家二位代码        （每日版本提供）
	WC string // World continent                // 世界大洲代码        （每日版本提供）
}

func ip2long(ipstr string) (uint32, []byte, error) {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return 0, nil, fmt.Errorf("bad ip string:%v", ip)
	}
	ip = ip.To4()
	if ip == nil {
		return 0, nil, fmt.Errorf("ipv6 not supported for now:%v", ip)
	}
	return binary.BigEndian.Uint32(ip), ip, nil
}

func Init(ipfile string) (ipip *Datx, err error) {
	ipip = new(Datx)
	ipdata, e := ioutil.ReadFile(ipfile)
	if e != nil {
		return
	}

	indexlenth := binary.BigEndian.Uint32(ipdata[:4])
	ipip.data = ipdata
	ipip.index = ipdata[4 : indexlenth+4]
	ipip.flag = make([]uint32, 65536)
	begin := 0
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			ipip.flag[i*256+j] = binary.LittleEndian.Uint32(ipip.index[begin : begin+4])
			begin += 4
		}
	}
	ipip.offset = indexlenth
	return
}

func (ipip *Datx) Find(ip string) (*IPIP, error) {
	ipLong, ips, err := ip2long(ip)
	if err != nil {
		return nil, err
	}
	ipPrefix := uint32(ips[0])*256 + uint32(ips[1])
	start := ipip.flag[ipPrefix]
	maxCompLen := ipip.offset - 262144 - 4
	var (
		indexOffset uint32 = 0
		indexLength uint32 = 0
	)

	for start := start*9 + 262144; start < maxCompLen; start += 9 {
		if binary.BigEndian.Uint32(ipip.index[start:start+4]) >= ipLong {
			indexOffset = binary.LittleEndian.Uint32(ipip.index[start+4 : start+8])
			indexLength = uint32((ipip.index[start+7] << 8) + ipip.index[start+8])
			break
		}
	}
	if indexOffset == 0 {
		return nil, fmt.Errorf("not found")
	}
	offset := ipip.offset + indexOffset - 262144
	result := ipip.data[offset : offset+indexLength]
	fields := bytes.Split(result, []byte("\t"))
	if len(fields) != 13 {
		return nil, fmt.Errorf("unexpected error")
	}
	return &IPIP{
		CR: string(fields[0]),
		RG: string(fields[1]),
		CT: string(fields[2]),
		CY: string(fields[3]),
		IS: string(fields[4]),
		LA: string(fields[5]),
		LN: string(fields[6]),
		T1: string(fields[7]),
		T2: string(fields[8]),
		AC: string(fields[9]),
		PC: string(fields[10]),
		CC: string(fields[11]),
		WC: string(fields[12]),
	}, nil
}
