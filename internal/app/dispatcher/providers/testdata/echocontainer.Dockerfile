# Used for integration testing Azure Batch provider, simply echo back the args passed in
FROM busybox
ENTRYPOINT ["echo"]