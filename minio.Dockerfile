# copy and paste this dockerfile into caprover to manually deploy
FROM minio/minio:latest

CMD ["server", "/data", "--console-address", ":9001"]