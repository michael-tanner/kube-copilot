# KubeCopilot.net

AI client for Kubernetes

## Features

- AI-powered CLI for Kubernetes management
- Natural language command support

## Getting Started

1. **Clone the repository:**
   ```sh
   git clone https://github.com/michael-tanner/kube-copilot.git
   cd kube-copilot
   ```

2. **Build the project:**
   ```sh
   go build -o kc
   ```

3. **Run the CLI:**
   ```sh
   ./kc "Get a list of namespaces and then for each namespace show a list of running pods. Return this in outline format"
   ```

4. **Example response:**
   ```sh
   Prompt: Get a list of namespaces and then for each namespace show a list of running pods. Return this in outline format
   Sending prompt to AI chat session...

   ..
   AI Function Call: Get the list of all namespaces in the cluster
   ......
   AI Function Call: Get running pods in the 'default' namespace

   AI Function Call: Get running pods in the 'kube-node-lease' namespace

   AI Function Call: Get running pods in the 'kube-public' namespace

   AI Function Call: Get running pods in the 'kube-system' namespace
   ........
   Here is the outline format of namespaces and their running pods:

   1. Namespace: default
      - backend-cf88d7d47-8k4vt
      - backend-cf88d7d47-r7wst
      - backend-cf88d7d47-tdh27
      - frontend-7c9955c498-nlmgj
      - frontend-7c9955c498-wbmh8
      - michael-02
      - michael-pod-2025

   2. Namespace: kube-node-lease
      - (No running pods)

   3. Namespace: kube-public
      - (No running pods)

   4. Namespace: kube-system
      - coredns-ccb96694c-8snnj
      - local-path-provisioner-5cf85fd-lpml4
      - metrics-server-5985cb-6m7rg
      - svclb-traefik-f334f-9nm9s
      - traefik-5d45fc-27r8k
   ```


5. **Requirements:**
   - Access to a Kubernetes cluster (e.g., Minikube or k3d)
   - Kube Copilot currently works best when run from within devcontainer of this repo

## Development

- This dev container supports Go, kubectl, Helm, Minikube, node, npm, and eslint pre-installed.
- Use the provided devcontainer for a ready-to-code environment.

## License

TBD
