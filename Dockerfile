FROM alpine

# We need to add the CA certs as otherwise Alpine won't trust anyone :(
RUN apk --no-cache add ca-certificates

# Copy our binary in
COPY jqd /bin/

# Open up port 9090
EXPOSE 9090

# Run it
CMD [ "/bin/jqd" ]
