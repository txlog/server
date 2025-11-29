# How to Deploy to Kubernetes

This guide provides a basic manifest to deploy Txlog Server to a Kubernetes cluster.

## Prerequisites

- A running Kubernetes cluster.
- `kubectl` configured.
- A PostgreSQL database accessible from the cluster.

## Deployment Manifest

Save the following content as `txlog-server.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: txlog-server
spec:
  replicas: 2
  selector:
    matchLabels:
      app: txlog-server
  template:
    metadata:
      labels:
        app: txlog-server
    spec:
      containers:
      - name: txlog-server
        image: cr.rda.run/txlog/server:main
        ports:
        - containerPort: 8080
        env:
        - name: INSTANCE
          value: "Production Cluster"
        - name: LOG_LEVEL
          value: "INFO"
        - name: PGSQL_HOST
          value: "postgres-service" # Replace with your DB host
        - name: PGSQL_PORT
          value: "5432"
        - name: PGSQL_USER
          value: "txlog"
        - name: PGSQL_DB
          value: "txlog"
        - name: PGSQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: txlog-secrets
              key: db-password
        # Add OIDC/LDAP env vars here if needed
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 15
---
apiVersion: v1
kind: Service
metadata:
  name: txlog-server
spec:
  selector:
    app: txlog-server
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: ClusterIP
```

## Steps to Deploy

1. **Create the Secret** for the database password:

    ```bash
    kubectl create secret generic txlog-secrets --from-literal=db-password='your_actual_password'
    ```

2. **Apply the Manifest**:

    ```bash
    kubectl apply -f txlog-server.yaml
    ```

3. **Verify Deployment**:

    ```bash
    kubectl get pods -l app=txlog-server
    ```

## Exposing the Service

To access the server from outside the cluster, you will typically use an Ingress. Here is a sample Ingress resource:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: txlog-ingress
spec:
  rules:
  - host: txlog.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: txlog-server
            port:
              number: 80
```
