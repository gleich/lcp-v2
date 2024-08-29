# copy and paste this dockerfile into caprover to manually deploy
FROM quay.io/minio/minio

CMD ["server", "/data", "--console-address", ":9001"]