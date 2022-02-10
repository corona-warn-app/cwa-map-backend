<h1 align="center">
    Corona-Warn-App Map Backend
</h1>

<p align="center">
    <a href="https://github.com/corona-warn-app/cwa-map-backend/commits/" title="Last Commit"><img src="https://img.shields.io/github/last-commit/corona-warn-app/cwa-map-backend?style=flat"></a>
    <a href="https://github.com/corona-warn-app/cwa-map-backend/issues" title="Open Issues"><img src="https://img.shields.io/github/issues/corona-warn-app/cwa-map-backend?style=flat"></a>
    <a href="https://github.com/corona-warn-app/cwa-map-backend/blob/master/LICENSE" title="License"><img src="https://img.shields.io/badge/License-Apache%202.0-green.svg?style=flat"></a>
</p>

<p align="center">
  <a href="#development">Development</a> •
  <a href="#documentation">Documentation</a> •
  <a href="#support-and-feedback">Support</a> •
  <a href="#how-to-contribute">Contribute</a> •
  <a href="#contributors">Contributors</a> •
  <a href="#repositories">Repositories</a> •
  <a href="#licensing">Licensing</a>
</p>

The goal of this project is providing an interface to the users to find COVID-19 testcenters in their region.

## About this component

The cwa-map-backend component provides the api for searching testcenters by their attributes, including location or test capabilities.

## Development
This component can be locally build in order to test the functionality of the interfaces and verify the concepts it is built upon.

### Prerequisites
- [Golang](https://go.dev/)
- *(optional)*: [Docker](https://www.docker.com)

### Build
After first checkout you have to install required dependencies by running
```bash
go build ./...
```

You can then access the frontend via http://localhost:9090

#### Docker based build
We recommend that you first check to ensure that [Docker](https://www.docker.com) is installed on your machine.

On the command line do the following:
```bash
docker build -t <imagename> .
docker run -p 8080:8080/tcp -it <imagename>
```

if you are in the root of the checked out repository.  
The docker image will then run on your local machine on port 8080 assuming you configured docker for shared network mode.

#### Remarks
This repository contains files which support our CI/CD pipeline and will be removed without further notice
- DockerfileCi - used for the GitHub build chain
- Jenkinsfile - used for Telekom internal SBS (**S**oftware**B**uild**S**ervice)

## Documentation
The full documentation for the Corona-Warn-App can be found in the [cwa-documentation](https://github.com/corona-warn-app/cwa-documentation) repository. The documentation repository contains technical documents, architecture information, and white papers related to this implementation.

## Support and feedback
The following channels are available for discussions, feedback, and support requests:

| Type                     | Channel                                                |
| ------------------------ | ------------------------------------------------------ |
| **General Discussion**   | <a href="https://github.com/corona-warn-app/cwa-documentation/issues/new/choose" title="General Discussion"><img src="https://img.shields.io/github/issues/corona-warn-app/cwa-documentation/question.svg?style=flat-square"></a> </a>   |
| **Concept Feedback**    | <a href="https://github.com/corona-warn-app/cwa-documentation/issues/new/choose" title="Open Concept Feedback"><img src="https://img.shields.io/github/issues/corona-warn-app/cwa-documentation/architecture.svg?style=flat-square"></a>  |
| **Map Service Issues**    | <a href="https://github.com/corona-warn-app/cwa-map-backend/issues" title="Open Issues"><img src="https://img.shields.io/github/issues/corona-warn-app/cwa-map-backend?style=flat"></a>  |
| **Other requests**    | <a href="mailto:opensource@telekom.de" title="Email CWA Team"><img src="https://img.shields.io/badge/email-CWA%20team-green?logo=mail.ru&style=flat-square&logoColor=white"></a>   |

## How to contribute
Contribution and feedback is encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](./CONTRIBUTING.md). By participating in this project, you agree to abide by its [Code of Conduct](./CODE_OF_CONDUCT.md) at all times.

## Contributors
The German government has asked SAP AG and Deutsche Telekom AG to develop the Corona-Warn-App for Germany as open source software. Deutsche Telekom is providing the network and mobile technology and will operate and run the backend for the app in a safe, scalable and stable manner. SAP is responsible for the app development, its framework and the underlying platform. Therefore, development teams of SAP and Deutsche Telekom are contributing to this project. At the same time our commitment to open source means that we are enabling -in fact encouraging- all interested parties to contribute and become part of its developer community.

## Repositories

A list of all public repositories from the Corona-Warn-App can be found [here](https://github.com/corona-warn-app/cwa-documentation/blob/master/README.md#repositories).

## Licensing
Copyright (c) 2020-2022 Deutsche Telekom AG.

Licensed under the **Apache License, Version 2.0** (the "License"); you may not use this file except in compliance with the License.

You may obtain a copy of the License at https://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the [LICENSE](./LICENSE) for the specific language governing permissions and limitations under the License.
