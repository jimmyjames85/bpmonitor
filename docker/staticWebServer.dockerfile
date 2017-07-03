# http://www.cicoria.com/simple-static-web-server-for-docker/


FROM    mhart/alpine-node

RUN     npm install -g http-server

WORKDIR /site

# The default port of the application
EXPOSE  8080
EXPOSE  443

CMD ["http-server", "--cors", "-p8080", "/site"]
