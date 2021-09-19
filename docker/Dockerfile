
FROM daprio/daprd AS dapr

COPY ./daprproxy /

EXPOSE 3500
EXPOSE 3000

CMD ["/daprproxy", "--http-dapr-port", "3500"]


