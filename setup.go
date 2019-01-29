package orchestrator

import (
	"fmt"
	"runtime"

	"github.com/netdata/go-orchestrator/pkg/multipath"
)

func (o *Orchestrator) Setup() bool {
	if o.Name == "" {
		log.Critical("name not set")
		return false
	}
	if o.Option == nil {
		log.Critical("cli options not set")
		return false
	}
	if len(o.Option.ConfigDir) != 0 {
		o.ConfigPath = multipath.New(o.Option.ConfigDir...)
	}
	if len(o.ConfigPath) == 0 {
		log.Critical("config path not set or empty")
		return false
	}
	if len(o.Registry) == 0 {
		log.Critical("registry not set or empty")
		return false
	}

	if o.configName == "" {
		o.configName = o.Name + ".conf"
	}
	configFile, err := o.ConfigPath.Find(o.configName)
	if err != nil {
		log.Critical("find config file error: ", err)
		return false
	}

	if err := loadYAML(o.Config, configFile); err != nil {
		log.Critical("loadYAML config error: ", err)
		return false
	}

	if !o.Config.Enabled {
		log.Info("disabled in configuration file")
		_, _ = fmt.Fprintln(o.Out, "DISABLE")
		return false
	}

	isAll := o.Option.Module == "all"
	for name, creator := range o.Registry {
		if !isAll && o.Option.Module != name {
			continue
		}
		if isAll && creator.DisabledByDefault && !o.Config.isModuleEnabled(name, true) {
			log.Infof("module '%s' disabled by default", name)
			continue
		}
		if isAll && !o.Config.isModuleEnabled(name, false) {
			log.Infof("module '%s' disabled in configuration file", name)
			continue
		}
		o.modules[name] = creator
	}

	if len(o.modules) == 0 {
		log.Critical("no modules to run")
		return false
	}

	if o.Config.MaxProcs > 0 {
		log.Infof("maximum number of used CPUs set to %d", o.Config.MaxProcs)
		runtime.GOMAXPROCS(o.Config.MaxProcs)
	} else {
		log.Infof("maximum number of used CPUs %d", runtime.NumCPU())
	}

	log.Infof("minimum update every %d", o.Option.UpdateEvery)

	return true
}
