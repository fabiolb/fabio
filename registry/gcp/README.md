### gcp - registry backend for routing to VM instances on the Google Cloud Platform
	
## Example 
This creates a container-vm running a fabio docker image stored in an image repository. Because it is a public facing router, the containerPort is mapped to 80. The admin is available on 8080.

### fabio-container.yaml

	version: v1
	kind: Pod
	metadata:
	  name: fabio-container
	spec:
	  containers:
	    - name: fabio-container
	      image: gcr.io/<yourrepo>/fabio:22
	      imagePullPolicy: IfNotPresent
	      volumeMounts:
	        - name: cacerts
	          mountPath: /etc/ssl/certs
	          readOnly: true
	      ports:
	        - containerPort: 9999
	          hostPort: 80
	          protocol: TCP
	        - containerPort: 9998
	          hostPort: 8080
	          protocol: TCP		
	      env:
	        - name: GCP_PROJECT
	          value: <yourproject>
	        - name: GCP_ZONE
	          value: europe-west1-b
	  restartPolicy: Always
	  dnsPolicy: Default
	  volumes:
	    - name: cacerts
	      hostPath:
	        path: /etc/ssl/certs	        	          
	        
### Create a fabio container

	gcloud compute instances create fabio-node-017 \
	  --image container-vm \
	  --metadata-from-file google-container-manifest=fabio-containers.yaml \
	  --zone europe-west1-b \
	  --machine-type n1-standard-1 \
	  --scopes cloud-platform	 
	 
### Create a traffic receiving vm instance

	gcloud compute instances create <yourinstancename> \
	  --tags fabio \
	  --image container-vm \
	  --metadata fabio="src=http://<yourhostname>/&dst=http://0.0.0.0:80/" \
	  --metadata-from-file google-container-manifest=<your-containers>.yaml \
	  --zone europe-west1-b \
	  --machine-type n1-standard-1 \
	  --scopes bigquery,cloud-platform,storage-full
	  
Note that `fabio` must be one of the tags because the gcp will filter instances based on this.
The `0.0.0.0` address in the destination part will be resolved by the gcp backend automatically.
The metadata 	is explained in the next section.

## Metadata

Meta data related to fabio provided by compute instances must use the following format (URL):

	src=http://inbound.com:80/incoming&dst=http://0.0.0.0:8080/out&weight=0.1&tags=v1
	
Mandatory keys
	
	src - scheme,host,port,path combination to match incoming fabio requests
	dst - scheme,host,port,path combination to forward incoming requests to downstream services
	
Optional keys

	weight - float number <0..1.0] for traffic splitting
	tags - comma separated list of labels
	
	
## Dockerfile
This was used to create the fabio:22 docker image in the example above.

	FROM busybox
	ADD fabio.properties /etc/fabio/fabio.properties
	ADD fabio /
	CMD ["/fabio", "-cfg", "/etc/fabio/fabio.properties"]	