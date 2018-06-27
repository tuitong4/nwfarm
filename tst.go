package main

import (
	"fmt"
	"regexp"
)

type TEST struct {
	Message string
}

func (t TEST) Say() {
	fmt.Println(t.Message)
}

func (t TEST) Update(m string) {
	t.Message = m
}

func (t *TEST) SayP() {
	fmt.Println(t.Message)
}

func (t *TEST) UpdateP(m string) {
	t.Message = m
}

func ParseSoftVersion_HUAWEI(v string) {
	re_v, _ := regexp.Compile(`V(\d+)R(\d+)C(\d+)SPC(\d+)`)
	re_p, _ := regexp.Compile(`V(\d+)R(\d+)SPH(\d+)`)
	ver  := re_v.FindStringSubmatch(v)
	path := re_p.FindStringSubmatch(v)
	fmt.Println(ver[1:], path[1:])
}

func main() {

	ver_string := `Huawei Versatile Routing Platform Software
VRP (R) software, Version 8.150 (CE12800 V200R002C50SPC800)
Copyright (C) 2012-2017 Huawei Technologies Co., Ltd.
HUAWEI CE12804 uptime is 281 days, 5 hours, 42 minutes 
Patch Version: V200R002SPH010

BKP  version information:
1.PCB      Version  : DE01BAK04A VER C
2.Board    Type     : CE-BAK04A  
3.MPU Slot Quantity : 2
4.LPU Slot Quantity : 4
5.SFU Slot Quantity : 6

MPU(Master) 5 : uptime is  281 days, 5 hours, 41 minutes
        StartupTime 2017/09/20   01:46:14+08:00 
Memory     Size     : 8192 M bytes
Flash      Size     : 4096 M bytes
NVRAM      Size     : 512 K bytes
MPU version information:                              
1.PCB      Version  : DE01MPUA VER C
2.MAB      Version  : 1 
3.Board    Type     : CE-MPUA
4.CPLD1    Version  : 104
5.BIOS     Version  : 386
                
MPU(Slave) 6 : uptime is  281 days, 5 hours, 41 minutes
        StartupTime 2017/09/20   01:46:20+08:00 
Memory     Size     : 8192 M bytes
Flash      Size     : 4096 M bytes
NVRAM      Size     : 512 K bytes
MPU version information:                              
1.PCB      Version  : DE01MPUA VER C
2.MAB      Version  : 1 
3.Board    Type     : CE-MPUA
4.CPLD1    Version  : 104
5.BIOS     Version  : 386
                
LPU 1 : uptime is 281 days, 5 hours, 32 minutes
        StartupTime 2017/09/20   01:55:39+08:00 
Memory     Size     : 4096 M bytes
Flash      Size     : 128  M bytes
LPU version information:
1.PCB      Version  : CEL36LQFD VER A
2.MAB      Version  : 1
3.Board    Type     : CE-L36LQ-FD
4.CPLD1    Version  : 104
5.CPLD2    Version  : 104
6.BIOS     Version  : 105
                
LPU 2 : uptime is 281 days, 5 hours, 31 minutes
        StartupTime 2017/09/20   01:56:29+08:00 
Memory     Size     : 4096 M bytes
Flash      Size     : 128  M bytes
LPU version information:
1.PCB      Version  : CEL36LQFD VER A
2.MAB      Version  : 1
3.Board    Type     : CE-L36LQ-FD
4.CPLD1    Version  : 104
5.CPLD2    Version  : 104
6.BIOS     Version  : 105

LPU 3 : uptime is 281 days, 5 hours, 31 minutes
        StartupTime 2017/09/20   01:56:45+08:00 
Memory     Size     : 4096 M bytes
Flash      Size     : 128  M bytes
LPU version information:
1.PCB      Version  : CEL36LQFD VER A
2.MAB      Version  : 1
3.Board    Type     : CE-L36LQ-FD
4.CPLD1    Version  : 104
5.CPLD2    Version  : 104
6.BIOS     Version  : 105

LPU 4 : uptime is 281 days, 5 hours, 32 minutes
        StartupTime 2017/09/20   01:55:36+08:00 
Memory     Size     : 4096 M bytes
Flash      Size     : 128  M bytes
LPU version information:
1.PCB      Version  : CEL36LQFD VER A
2.MAB      Version  : 1
3.Board    Type     : CE-L36LQ-FD
4.CPLD1    Version  : 104
5.CPLD2    Version  : 104
6.BIOS     Version  : 105

SFU 9 : uptime is 281 days, 5 hours, 33 minutes
        StartupTime 2017/09/20   01:54:23+08:00 
Memory     Size     : 512 M bytes
Flash      Size     : 64  M bytes
SFU version information:
1.PCB      Version  : CESFU04G VER A
2.MAB      Version  : 1
3.Board    Type     : CE-SFU04G
4.CPLD1    Version  : 101
5.BIOS     Version  : 386

SFU 10 : uptime is 281 days, 5 hours, 33 minutes
        StartupTime 2017/09/20   01:54:23+08:00 
Memory     Size     : 512 M bytes
Flash      Size     : 64  M bytes
SFU version information:
1.PCB      Version  : CESFU04G VER A
2.MAB      Version  : 1
3.Board    Type     : CE-SFU04G
4.CPLD1    Version  : 101
5.BIOS     Version  : 386

SFU 11 : uptime is 281 days, 5 hours, 33 minutes
        StartupTime 2017/09/20   01:54:24+08:00 
Memory     Size     : 512 M bytes
Flash      Size     : 64  M bytes
SFU version information:
1.PCB      Version  : CESFU04G VER A
2.MAB      Version  : 1
3.Board    Type     : CE-SFU04G
4.CPLD1    Version  : 101
5.BIOS     Version  : 386
                
SFU 12 : uptime is 281 days, 5 hours, 33 minutes
        StartupTime 2017/09/20   01:54:23+08:00 
Memory     Size     : 512 M bytes
Flash      Size     : 64  M bytes
SFU version information:
1.PCB      Version  : CESFU04G VER A
2.MAB      Version  : 1
3.Board    Type     : CE-SFU04G
4.CPLD1    Version  : 101
5.BIOS     Version  : 386

CMU(Slave) 7 : uptime is 279 days, 21 hours, 57 minutes
        StartupTime 2017/09/20   02:01:35+08:00
Memory     Size     : 128 M bytes
Flash      Size     : 32  M bytes
CMU version information:
1.PCB      Version  : DE01CMUA VER B
2.MAB      Version  : 1 
3.Board    Type     : CE-CMUA
4.CPLD1    Version  : 104
5.BIOS     Version  : 127

CMU(Master) 8 : uptime is 279 days, 13 hours, 46 minutes
        StartupTime 2017/09/20   01:58:06+08:00
Memory     Size     : 128 M bytes
Flash      Size     : 32  M bytes
CMU version information:
1.PCB      Version  : DE01CMUA VER B
2.MAB      Version  : 1 
3.Board    Type     : CE-CMUA
4.CPLD1    Version  : 104
5.BIOS     Version  : 127`

ParseSoftVersion_HUAWEI(ver_string)
}
