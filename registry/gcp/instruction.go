package gcp

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"

	"google.golang.org/api/compute/v1"
)

// src=http://inbound.com:80/foo&dst=http://0.0.0.0:8080/bar&weight=0.1&tags=v1
// route add {instance.name} inbound.com/foo http://{instance.networkIP}:8080/bar tags=v1,"{instance.tags.items}" weight=0.1
func buildInstruction(instance *compute.Instance, spec string) string {
	v, err := url.ParseQuery(spec)
	if err != nil {
		log.Printf("[ERROR] instance %s has invalid fabio spec %v", instance.Name, err)
		return ""
	}
	src, err := url.Parse(v.Get("src"))
	if err != nil {
		log.Printf("[ERROR] invalid src url:%v", err)
		return ""
	}
	if !validate(instance.Name, "src.host", src.Host) {
		return ""
	}
	host := src.Host
	path := src.Path
	if len(path) == 0 {
		path = "/"
	}
	if len(instance.NetworkInterfaces) == 0 {
		log.Printf("[ERROR] decoding fabio route parameters %v", err)
		return ""
	}

	dst, err := url.Parse(v.Get("dst"))
	if err != nil {
		log.Printf("[ERROR] invalid dst url:%v", err)
		return ""
	}
	if !validate(instance.Name, "dst.host", dst.Host) {
		return ""
	}
	hp := strings.Split(dst.Host, ":")
	ip := hp[0]
	targetPort := "80"
	// take port from dst if specfied
	if len(hp) == 2 {
		targetPort = hp[1]
	}
	// check for placeholder
	if "0.0.0.0" == ip {
		if len(instance.NetworkInterfaces) == 0 {
			log.Printf("[ERROR] missing network interfaces in instance description")
			return ""
		}
		ip = instance.NetworkInterfaces[0].NetworkIP
	}
	targetPath := dst.Path
	if len(targetPath) == 0 {
		targetPath = "/"
	}
	weight := v.Get("weight")

	out := new(bytes.Buffer)
	out.WriteString("route add ")
	out.WriteString(instance.Name)
	out.WriteString(" ")
	out.WriteString(host)
	out.WriteString(path)
	out.WriteString(" http://")
	out.WriteString(ip)
	out.WriteString(":")
	out.WriteString(targetPort)
	out.WriteString(targetPath)

	if len(weight) > 0 {
		fmt.Fprintf(out, " weight=%s", weight)
	}

	if instance.Tags != nil {
		tagCount := len(instance.Tags.Items)
		if tagCount > 0 {
			out.WriteString(" tags \"")
			for i, each := range instance.Tags.Items {
				if i > 0 {
					out.WriteString(",")
				}
				out.WriteString(each)
			}
			out.WriteString("\"")
		}
	}
	return out.String()
}

func validate(instanceName, param, value string) bool {
	if len(value) == 0 {
		log.Printf("[ERROR] instance %s is missing %s parameter for fabio route", instanceName, param)
		return false
	}
	return true
}
