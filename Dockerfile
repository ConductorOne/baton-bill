FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-bill"]
COPY baton-bill /