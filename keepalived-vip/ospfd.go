package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"text/template"

	"github.com/golang/glog"
)

const ospfdCfg = "/etc/ospfd/ospfd.conf"

var ospfdTmpl = "ospfd.tmpl"

type ospfd struct {
	linkip  string
	started bool
	tmpl    *template.Template
	cmd     *exec.Cmd
}

// WriteCfg create a new ospfd configuration file.
// In case of an error with the generation it returns the error
func (o *ospfd) WriteCfg(svcs []vip) error {
	w, err := os.Create(ospfdCfg)
	if err != nil {
		return err
	}
	defer w.Close()

	conf := make(map[string]interface{})
	conf["svcs"] = svcs
	conf["linkip"] = o.linkip

	addrmap := make(map[string][]string)

	for _, s := range svcs {
		addrmap[s.IP] = append(addrmap[s.IP], s.Name)
	}

	conf["addressmap"] = addrmap

	if glog.V(2) {
		b, _ := json.Marshal(conf)
		glog.Infof("%v", string(b))
	}

	return o.tmpl.Execute(w, conf)
}

func (o *ospfd) Start() {
	o.cmd = exec.Command("/etc/init.d/ospfd", "start")

	o.cmd.Stdout = os.Stdout
	o.cmd.Stderr = os.Stderr
	o.started = true

	if err := o.cmd.Start(); err != nil {
		glog.Errorf("ospfd error: %v", err)
	}

	if err := o.cmd.Wait(); err != nil {
		glog.Errorf("ospfd error: %v", err)
	}
}

func (o *ospfd) Restart() error {
	var err error
	o.cmd = exec.Command("/etc/init.d/ospfd", "restart")

	o.cmd.Stdout = os.Stdout
	o.cmd.Stderr = os.Stderr
	o.started = true

	if err = o.cmd.Start(); err != nil {
		glog.Errorf("ospfd error: %v", err)
	}

	if err = o.cmd.Wait(); err != nil {
		glog.Errorf("ospfd error: %v", err)
	}
	return err
}

func (o *ospfd) loadTemplate() error {
	tmpl, err := template.ParseFiles(ospfdTmpl)
	if err != nil {
		return err
	}
	o.tmpl = tmpl
	return nil
}
