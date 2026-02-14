## AVOXI Coding Challenge Details:

Time limit - 4 Hours

Scenario (note, this is fictional, but gives a sense of the types of requests we might encounter):

Our customers are requesting we restrict users logging in to their UI accounts from selected countries so that they can prevent them from outsourcing their work to others. For an initial phase one we’re not going to worry about VPN connectivity, only the presented IP address.

The team has designed a solution where the customer database will hold the white listed countries and the API gateway will capture the requesting IP address, check the target customer for restrictions, and send the data elements to a new service you are going to create.

The new service will be an HTTP-based API that receives an IP address and a list of allowed countries. The API should return an indicator if the IP address is within the allowed countries or not. You can get a data set of IP address to country mappings from https://dev.maxmind.com/geoip/geoip2/geolite2/.

We do our backend development in Go (Golang) and prefer solutions in that language but can accept solutions in any major programming language.

You are allowed to utilize AI coding tools to complete this coding challenge but are expected to provide usage details in submission notes.

We'll be explicitly looking at coding style, code organization, API design, and operational/maintenance aspects such as logging and error handling. We'll also be giving bonus points for things like

- Documenting a plan for keeping the mapping data up to date. Extra bonus points for implementing the solution.

- Including a Docker file for the running service

- Including a Kubernetes YAML file for running the service in an existing cluster

- Exposing the service as gRPC in addition to HTTP

- Other extensions to the service you think would be worthwhile. If you do so, please include a brief description of the feature and justification for its inclusion. Think of this as what you would have said during the design meeting to convince your team the effort was necessary.

We'd like you to spend no more than 4 hours working on the solution over the next two days. We can accept submissions in two mechanisms.

1. Our preferred mechanism is for you to create a public project in your personal Github, Bitbucket or similar service and send us a link.

2. Create a ZIP file and place it in a Google Drive, Dropbox, or other file sharing service and send us a link.
