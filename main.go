package main

import (
	"bytes"
	"fmt"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"os/exec"
	"strings"
	"flag"
)

var (
	h bool
	conf string
)

func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&conf,"config","config.ini","配置文件路径，默认是当前目录下的conf.ini文件")
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `muiltnet version:v1.0
Usage: muiltnet [-h] [-config filename] [run command]

Options:
`)
	flag.PrintDefaults()
}

func main() {
	var cfg *ini.File
	var err error
	flag.Parse()
	if h {
		flag.Usage()
		os.Exit(0)
	}

	if conf!=""{
		cfg, err = ini.Load(conf)
		if err != nil {
			log.Fatal("Fail to read file: ", err)
		}
	}else {
		cfg, err = ini.Load("conf.ini")
		if err != nil {
			log.Fatal("Fail to read file: ", err)
		}
	}
	lockCmd := "find / -name *ucloud*"
	command:=exec.Command("/bin/bash", "-c",lockCmd)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Run(); err != nil {
		log.Println(err)
		os.Exit(0)
	}
if len(stdout.String())<=0{
	log.Println("本程序目前只支持在UCloud云主机运行")
	os.Exit(0)
}

	childs:=cfg.Section("interface").ChildSections()
	if len(childs) >0 {
		for _,child :=range childs{
			name:=child.Key("name").String()
			ip:=child.Key("ip").String()
			mask:=child.Key("mask").String()
			gw:=child.Key("gw").String()
			run:=child.Key("run").String()
			if name !="" && ip !="" && mask !="" && gw !="" {
				//检查当前接口的是否创建好netns
				command:=exec.Command("/bin/bash", "-c","ip netns show")
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}
				command.Stdout = stdout
				command.Stderr = stderr
				if err := command.Run(); err != nil {
					log.Println(err)
					log.Println(stderr.String())
					os.Exit(0)
				}
				output:=stdout.String()
				if len(output)>0 {
					boolContain:=strings.Contains(output,name)
					if !boolContain {
						cmdAdd:="ip netns add "+name
						cmdSet:="ip link set "+name +" netns " +name
						cmdIp:="ip netns exec "+name +" ifconfig "+name +" " +ip +" netmask " + mask +" up"
						cmdMtu:="ip netns exec "+name + " ifconfig "+name +" mtu 1452"
						cmdGw:="ip netns exec "+name + " route add default gw "+ gw
						cmdSsh:="ip netns exec "+name + " /usr/sbin/sshd"
						cmd:=cmdAdd +" && "+cmdSet + " && " + cmdIp +" && "+ cmdMtu +" && "+ cmdGw+" && "+ cmdSsh
						command:=exec.Command("/bin/bash", "-c",cmd)
						stdout := &bytes.Buffer{}
						stderr := &bytes.Buffer{}
						command.Stdout = stdout
						command.Stderr = stderr
						if err := command.Run(); err != nil {
							log.Println(err)
							log.Printf(stderr.String())
							os.Exit(0)
						}
					}
				}else
				{
					cmdAdd:="ip netns add "+name
					cmdSet:="ip link set "+name +" netns " +name
					cmdIp:="ip netns exec "+name +" ifconfig "+name +" " +ip +" netmask " + mask +" up"
					cmdMtu:="ip netns exec "+name + " ifconfig "+name +" mtu 1452"
					cmdGw:="ip netns exec "+name + " route add default gw "+ gw
					cmdSsh:="ip netns exec "+name + " /usr/sbin/sshd"

					cmd:=cmdAdd +" && "+cmdSet + " && " + cmdIp +" && "+ cmdMtu +" && "+ cmdGw+" && "+ cmdSsh
					command:=exec.Command("/bin/bash", "-c",cmd)
					stdout := &bytes.Buffer{}
					stderr := &bytes.Buffer{}
					command.Stdout = stdout
					command.Stderr = stderr
					if err := command.Run(); err != nil {
						log.Println(err)
						log.Printf(stderr.String())
						os.Exit(0)
					}
				}
			}else {
				os.Exit(0)
			}
			if run !="" {
				runCmd:=strings.Split(run,",")
				for _,cmd :=range runCmd{
					fmt.Println("执行命令：" + cmd)
					netnsCmd := "nohup ip netns exec "+name +" " + cmd +" >/tmp/c.log 2>&1 &"
					fmt.Println(netnsCmd)
					command:=exec.Command("/bin/bash", "-c",netnsCmd)
					stdout := &bytes.Buffer{}
					stderr := &bytes.Buffer{}
					command.Stdout = stdout
					command.Stderr = stderr
					if err := command.Run(); err != nil {
						log.Println(err)
						log.Println(stderr.String())
						os.Exit(0)
					}
					log.Println(stdout.String())
				}
			}
		}
	}

}
