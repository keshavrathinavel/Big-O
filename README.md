
# 🏎️ Load Test and Database Generation Automation

This project uses a `Justfile` to automate the following tasks:  
✅ Building binaries
✅ Generating dataset
✅ Running load tests with configurable parameters
✅ Managing Docker containers

---

## ⚠️ **Requirements**
- `just` installed  
- **Go 1.24.1 or higher** installed  
- Docker installed and running  

---

## 🚀 **Setup**
### **1. Install `just`**
Install [`just`](https://github.com/casey/just) if not already installed:

- **macOS (Homebrew):**  
```bash
brew install just
```

---

### **2. Install Go**
> **Go 1.24.1 or higher is required**  
Install Go using the following commands:

- **macOS (Homebrew):**  
```bash
brew install go
```

---

### **3. Clone the Repository**
Clone the repository and navigate to the project folder:

```bash
git clone git@github.com-rohanjnr:keshavrathinavel/Big-O.git
cd Big-O
```

---

## 📄 **Commands Overview**

### **1. Build Binaries**
You can manually build the binaries before running other commands:

- **Build the `gen` binary**  
```bash
just build-gen
```

- **Build the `load_test` binary**  
```bash
just build-load-test
```

---

### **2. Generate Dataset**
Generate the database by building the `gen` binary and executing it:
```bash
just generate-database
```
- This will generate a database with:
  - **Number of entries** = `7150000 * 8`
  - **Parallelism** = `4`

---

### **3. Run Load Test**
You can run the load test with default or custom values:

**3.1 Setup config file**

Edit the existing config file with the server addresses. You can simulate this by running your BigO Solution in 7 coontainers or as 7 processes on different ports.

**3.2 Running the test**  
```bash
just load-test --reqs 1000000 --vus 2
```

- `reqs` → Number of requests per virtual user
- `vus` → Number of virtual users

**3.3 Capturing Metrics**

Metrics will be available on the grafana dashboard.

1. Visit http://localhost:3000/d/befi36fr71atca/bigo-monitoring
2. In the Reqs/Sec Graph, select the portion of the graph post-request rampup and pre-request ramp down (basically the first highest peak and the last highest peak). This can be done by left clicking and dragging the mouse across the two points.
3. Capture a screenshot containing the graphs in the dashboard.

This can be iterative processes to find the optimal number of virtual users (VU) for your solution.

The screenshot and the VU count needs to be updated in the registrations sheet, can be found in slack.

---

### **4. Stop Docker Containers**
Stop running Docker containers:
```bash
just stop-docker
```

---
