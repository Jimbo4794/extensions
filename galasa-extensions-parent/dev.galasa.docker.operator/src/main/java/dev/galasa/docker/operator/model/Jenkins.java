/*
 * Licensed Materials - Property of IBM
 * 
 * (c) Copyright IBM Corp. 2020.
 */
package dev.galasa.docker.operator.model;

import java.util.ArrayList;
import java.util.List;

import com.github.dockerjava.api.DockerClient;
import com.github.dockerjava.api.command.CreateContainerCmd;
import com.github.dockerjava.api.command.InspectContainerResponse;
import com.github.dockerjava.api.exception.NotFoundException;
import com.github.dockerjava.api.model.ExposedPort;
import com.github.dockerjava.api.model.HostConfig;
import com.github.dockerjava.api.model.Mount;
import com.github.dockerjava.api.model.MountType;
import com.github.dockerjava.api.model.PortBinding;
import com.github.dockerjava.api.model.Ports.Binding;
import com.github.dockerjava.api.model.Volume;

import dev.galasa.docker.operator.DockerOperatorException;
import dev.galasa.docker.operator.config.JenkinsConfiguration;

public class Jenkins extends AbstractContainerResource {

    private static final String RESOURCE_NAME = "galasa_jenkins";

    private final JenkinsVolume jenkinsVolume;

    public Jenkins(Ecosystem ecosystem) {
        super(ecosystem, RESOURCE_NAME);

        this.jenkinsVolume = new JenkinsVolume(ecosystem);
        this.jenkinsVolume.addDependency(this);
        ecosystem.addResource(this.jenkinsVolume);
    }

    @Override
    public void checkResourceDefined() throws DockerOperatorException {
        this.jenkinsVolume.checkResourceDefined();

        String jenkinsVolumeName = this.jenkinsVolume.getName();
        String targetImageName = getTargetImageName();

        DockerClient dockerClient = getEcosystem().getDockerClient();

        String imageId = null;
        try {
            imageId = getImageId(targetImageName);
        } catch(Exception e) {
            throw new DockerOperatorException("Problem determining Jenkins image id", e);
        }

        boolean found   = true;
        boolean correct = true;
        try {
            InspectContainerResponse response = dockerClient.inspectContainerCmd(RESOURCE_NAME).exec();

            String actualImage = response.getImageId();
            if (!imageId.equals(actualImage)) {
                correct = false;
            }

            List<com.github.dockerjava.api.command.InspectContainerResponse.Mount> mounts = response.getMounts();
            if (mounts == null || mounts.size() != 1) {
                correct = false;
            } else {
                if (!mounts.get(0).getName().equals(jenkinsVolumeName)) {
                    correct = false;
                } else {
                    Volume dir = mounts.get(0).getDestination();
                    if (!dir.getPath().equals("/var/jenkins_home")) {
                        correct = false;
                    }
                }
            }
        } catch(NotFoundException e) {
            found = false;
        } catch(Exception e) {
            throw new DockerOperatorException("Problem inspecting Jenkins container", e);
        }

        if (found && correct) {
            return;
        }

        for(AbstractResource dependency : getDependencies()) {
            dependency.dependencyChanging(DependencyEvent.DEFINING, this);
        }

        if (!correct) {
            stopContainer();
            deleteContainer();
        }

        try {
            System.out.println("Defining the Jenkins container");
            CreateContainerCmd cmd = dockerClient.createContainerCmd(RESOURCE_NAME);
            cmd.withName(RESOURCE_NAME);
            cmd.withImage(targetImageName);

            HostConfig hostConfig = new HostConfig();
            String portNumber = Integer.toString(getEcosystem().getConfiguration().getJenkins().getPort());
            hostConfig.withPortBindings(new PortBinding(new Binding("0.0.0.0", portNumber), new ExposedPort(8080)));

            Mount mount = new Mount();
            mount.withType(MountType.VOLUME);
            mount.withSource(jenkinsVolumeName);
            mount.withTarget("/var/jenkins_home");
            ArrayList<Mount> mounts = new ArrayList<>();
            mounts.add(mount);

            hostConfig.withMounts(mounts);
            cmd.withHostConfig(hostConfig);

            cmd.exec();
        } catch(Exception e) {
            throw new DockerOperatorException("Problem creating Jenkins container", e);
        }
    }


    @Override
    protected void checkResourceRunning() throws DockerOperatorException {
        checkContainerRunning();
    }


    protected void checkContainerRunning() throws DockerOperatorException {
        if (isContainerRunning(false)) {
            return;
        }
        System.out.println("Jenkins Container is down, requires start up");

        for(AbstractResource dependency : getDependencies()) {
            dependency.dependencyChanging(DependencyEvent.STARTING, this);
        }

        try {
            System.out.println("Starting Jenkins container");
            startContainer();
        } catch(Exception e) {
            throw new DockerOperatorException("Problem starting Nexus container", e);
        }

        try {
            checkLog("Jenkins is fully up and running", 300);
            System.out.println("Jenkins container is up");
        } catch(Exception e) {
            System.out.println("Failed to detect Jenkins up message, deleting container to force rebuild");
            deleteContainer();
            throw new DockerOperatorException("Problem waiting for Jenkins container started message", e);
        }

    }



    @Override
    protected void dependencyChanging(DependencyEvent event, AbstractResource resource) throws DockerOperatorException {
        switch(event) {
            case DEFINING:
                stopContainer();
                break;
            default:
                throw new DockerOperatorException("Unexpected event '" + event + " from dependency " + resource.getClass().getName());
        }
    }


    private String getTargetImageName() {
        JenkinsConfiguration jenkinsConfig = getEcosystem().getConfiguration().getJenkins();
        return jenkinsConfig.getImage() + ":" + jenkinsConfig.getVersion();
    }

}
