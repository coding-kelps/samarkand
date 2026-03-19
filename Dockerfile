FROM gcr.io/distroless/static-debian13:nonroot AS runner

ARG TARGETPLATFORM

COPY --chown=nonroot:nonroot $TARGETPLATFORM/samarkand /bin/

EXPOSE 8080 50051

ENTRYPOINT ["/bin/samarkand"]

CMD ["start"]
