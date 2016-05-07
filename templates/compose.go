package templates

// SupportServicesCompose is the compose.yml file for the dev support services.
var SupportServicesCompose = []byte(`nginx:
  image: jwilder/nginx-proxy
  ports:
    - "80:80"
  volumes:
    - /var/run/docker.sock:/tmp/docker.sock:ro
  restart: always
dnsmasq:
  image: andyshinn/dnsmasq
  ports:
    - "127.0.0.1:53:53/tcp"
    - "127.0.0.1:53:53/udp"
  cap_add:
    - NET_ADMIN
  command: --address=/dev/127.0.0.1
  restart: always`)
