package code

import (
	"fmt"

	"github.com/jcelliott/lumber"
	"github.com/nanobox-io/golang-docker-client"

	container_generator "github.com/nanobox-io/nanobox/generators/containers"
	hook_generator "github.com/nanobox-io/nanobox/generators/hooks/code"
	"github.com/nanobox-io/nanobox/models"
	"github.com/nanobox-io/nanobox/util/dhcp"
	"github.com/nanobox-io/nanobox/util/display"
	"github.com/nanobox-io/nanobox/util/provider"
	"github.com/nanobox-io/nanobox/util/hookit"
)

//
func Setup(appModel *models.App, componentModel *models.Component, warehouseConfig WarehouseConfig) error {
	display.OpenContext("setting up %s", componentModel.Name)
	defer display.CloseContext()

	// generate the missing component data
	if err := componentModel.Generate(appModel, "code"); err != nil {
		lumber.Error("component:Setup:models.Component:Generate(%s, code): %s", appModel.ID, componentModel.Name, err.Error())
		return fmt.Errorf("failed to generate component data: %s", err.Error())
	}

	// short-circuit if this component is already setup
	if componentModel.State != "initialized" {
		return nil
	}

	// generate a docker percent display
	dockerPercent := &display.DockerPercentDisplay{
		Output: display.NewStreamer("info"),
		Prefix: componentModel.Image,
	}

	// pull the component image
	display.StartTask("pulling %s image", componentModel.Image)
	if _, err := docker.ImagePull(componentModel.Image, dockerPercent); err != nil {
		lumber.Error("component:Setup:docker.ImagePull(%s, nil): %s", componentModel.Image, err.Error())
		display.ErrorTask()
		return fmt.Errorf("failed to pull docker image (%s): %s", componentModel.Image, err.Error())
	}
	display.StopTask()
	//

	if err := reserveIps(componentModel); err != nil {
		lumber.Error("code:Setup:setup.getLocalIP(): %s", err.Error())
		return err
	}

	// create docker container
	display.StartTask("starting container")
	config := container_generator.ComponentConfig(componentModel)
	container, err := docker.CreateContainer(config)
	if err != nil {
		lumber.Error("code:Setup:createContainer:docker.CreateContainer(%+v): %s", config, err.Error())
		display.ErrorTask()
		return err
	}
	display.StopTask()

	// save the component
	componentModel.ID = container.ID
	if err := componentModel.Save(); err != nil {
		lumber.Error("code:Setup:Component.Save(): %s", err.Error())
		return err
	}

	//
	if err := addIPToProvider(componentModel); err != nil {
		return err
	}

	lumber.Prefix("code:Setup")
	defer lumber.Prefix("")

	// run fetch build command
	fetchPayload := hook_generator.FetchPayload(componentModel, warehouseConfig.WarehouseURL)

	display.StartTask("fetching code")
	if _, err := hookit.RunFetchHook(componentModel.ID, fetchPayload); err != nil {
		display.ErrorTask()
		return err
	}
	display.StopTask()

	// run configure command
	payload := hook_generator.ConfigurePayload(appModel, componentModel)

	//
	display.StartTask("configuring code")
	if _, err := hookit.RunConfigureHook(componentModel.ID, payload); err != nil {
		display.ErrorTask()
		return fmt.Errorf("failed to configure code: %s", err.Error())
	}
	display.StopTask()

	// run start command
	display.StartTask("starting code")
	if _, err := hookit.RunStartHook(componentModel.ID, payload); err != nil {
		display.ErrorTask()
		return err
	}
	display.StopTask()

	//
	componentModel.State = ACTIVE
	if err := componentModel.Save(); err != nil {
		lumber.Error("code:Configure:Component.Save(): %s", err.Error())
		return fmt.Errorf("unable to save component model: %s", err.Error())
	}

	return nil
}

//  ...
func reserveIps(componentModel *models.Component) error {

	if componentModel.InternalIP == "" {
		localIP, err := dhcp.ReserveLocal()
		if err != nil {
			lumber.Error("code:Setup:dhcp.ReserveLocal(): %s", err.Error())
			return err
		}
		componentModel.InternalIP = localIP.String()
	}

	if componentModel.ExternalIP == "" {
		ip, err := dhcp.ReserveGlobal()
		if err != nil {
			lumber.Error("code:Setup:dhcp.ReserveGlobal(): %s", err.Error())
			return err
		}
		componentModel.ExternalIP = ip.String()
	}

	return nil
}

// createContainer ...
func createContainer(componentModel *models.Component) error {
	display.StartTask("creating container")

	display.StopTask()

	return nil
}

// addIPToProvider ...
func addIPToProvider(componentModel *models.Component) error {
	display.StartTask("building network")
	if err := provider.AddIP(componentModel.ExternalIP); err != nil {
		display.ErrorTask()
		lumber.Error("code:Setup:addIPToProvider:provider.AddIP(%s): %s", componentModel.ExternalIP, err.Error())
		return err
	}

	if err := provider.AddNat(componentModel.ExternalIP, componentModel.InternalIP); err != nil {
		lumber.Error("code:Setup:addIPToProvider:provider.AddNat(%s, %s): %s", componentModel.ExternalIP, componentModel.InternalIP, err.Error())
		display.ErrorTask()
		return err
	}
	display.StopTask()
	return nil
}