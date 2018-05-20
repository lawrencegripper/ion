![](docs/ion-logo.png)

[![Build Status](https://travis-ci.org/lawrencegripper/ion.svg?branch=master)](https://travis-ci.org/lawrencegripper/ion)

> **Warning**: Ion is currently under initial development - please **do not** use in production and have patience whilst we flesh things out!

**Ion** is a scalable, event-driven, data processing platform. It allows you to create complex workflows as a graph of loosley coupled tasks. You define your control logic using programming languages so you are not constrained by a DSL or markdown language.

## Overview
Imagine you wish to derive insights from a large collection of videos hosted on [YouTube](https://www.youtube.com). Let's say you you want to find and identifying certain types of car present in each of the videos. You could write a script to grab each of the video URLs and post it to Ion's HTTP gateway. The first job could parse the URL and go and download the video from YouTube, sample it and store the sampled video in Ion's data plane. As well as storing the video, the job raises a new `new_video` event which it too publishes via Ion's data plane. A subsequent job that is subscribed to the `new_video` messaging topic will then be dispatched, this job might run some object detection algorithms over the sampled video to try and detect cars. Once it finds a car in a frame, it could crop it out and write the cropped image to Ion's date plane before raising a `car_detected` event. This event could be picked up by another job that takes the cropped images of the car and runs a more specialised car classification algorithm over it to work out the car's make, model and colour. These insights can then too be written and persited into Ion's dataplane for future analysis. You might then write a web front end which queries Ion's data plane to allow users to query for videos containing certain car types.
This pattern of passing data between multiple levels of analysis can be applied to many use cases and problems. Ion is a generic framework that allows you to run a workflow of jobs whilst handling the task of migrating the required data between those jobs for you.

![](docs/ion.PNG)

## Goals
* Scalable
* Flexible
* Extensible
* Simple

## Blueprint
Ion is built to harness the power of cloud platform services that allow it to be elastically scalable, fault tolerant and automatically managed.

The Ion platform is comprised of 4 main components:
1. [Dispatchers](#dispatchers)
2. [Jobs](#jobs)
3. [Modules](#modules)
4. [Handlers](#handlers)

### Dispatchers
Dispatchers can be configured to subscribe to certain event topics on an external messaging service. Multiple Dispatchers can subscribe for the same events to allow scalability. When a Dispatcher dequeues a message, it extracts requirements about the job, applies business logic rules and then dispatches a new job to be scheduled onto an appropriate execution environment.
Supported executors include:
* [Kubernetes](https://kubernetes.io/)
* [Azure Batch](https://azure.microsoft.com/en-us/services/batch/)
* [Azure Container Instances (ACI)](https://azure.microsoft.com/en-us/services/container-instances/) - Coming Soon
* Local

The Dispatcher uses `providers` to schedule jobs onto execution environments. If you wish to add support for a new executor, please consider writing a custom provider and submitting a PR.

For more details on dispatchers: please refer to the [Dispatcher docs](dispatcher/README.md)

### Jobs
A job is the unit of deployment for a Dispatcher, it contains at least 2 sub components; a module and a handler. The job should be provided with all the parameters it needs to successfully fulfil its work during dispatch.

### Modules
A module is analogous to a discrete task. It is the program that the user wishes to execute in response to a particular event. This is likely the only component a user of the platform needs to be concerned with. The module can be written in any language but should be containerized along with all its dependencies.

For more details on modules: please refer to the [Module docs](modules/README.md)

### Handlers
Each module is executed as part of a series of containers. There is a prepare handler container ran first, then the module runs, then finally a commit handler container is run.
The prepare handler is responsible for creating the environment in which the module will run. Including creating the desired directory structure and populating it with any data passed from previous modules.
The commit handler is responsible for persisting any state written out by a module to Ion's data plane.

For more details on handler: please refer to the [Handler docs](handler/README.md)

## Developing

Ensure you have go setup correctly then run `go get github.com/lawrencegripper/ion` to pull the source into your gopath

To check changes using the same process as the CI build run `./ci.sh` at the root directory and check the output for any errors.

This will check the `/dispatcher` and `/handler` directories. If you add an additional folder/service add your new path and add a new line into `./ci.sh` to invoke it with the correct `FOLDER` build arg.