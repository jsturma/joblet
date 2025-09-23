# ADR-008: GPU Support Architecture

## Status

Proposed

## Context

Here's the thing - more and more of our users are asking about GPU support. It makes sense. If you're training a neural
network or crunching through massive datasets, CPUs just don't cut it anymore. You need that GPU acceleration to get
anything done in reasonable time.

But adding GPU support to Joblet isn't as simple as just passing through a device. We've built this really solid
isolation model using native Linux features - namespaces and cgroups - no Docker, no containers, just pure kernel
capabilities. We can't just punch holes in it for GPUs. That would defeat the whole purpose of what makes Joblet
special.

The tricky part is that GPUs come with their own baggage. You've got NVIDIA drivers that need kernel modules, CUDA
libraries that have to match specific versions, and these GPUs cost a fortune so you can't just dedicate one to each
job. Plus, GPU memory is this whole separate thing from system RAM - you can't just use our existing memory limits.

And then there's the multi-tenancy problem. When you're sharing a $10,000 GPU between multiple jobs, you really need to
make sure one job can't peek at another job's GPU memory. That's a security nightmare waiting to happen

## Decision

After a lot of back and forth, we've decided to go all-in on GPU support, but we're doing it the Joblet way - using
native Linux kernel features, no container runtimes, no extra layers of abstraction. Just like we use namespaces and
cgroups directly for CPU and memory isolation, we'll use the kernel's device control capabilities for GPUs.

We're starting with NVIDIA because, let's face it, that's what everyone uses for compute. AMD is getting better, but
NVIDIA owns this space right now. We'll design it so we can add other vendors later, but NVIDIA is the priority.

### How We're Thinking About This

First off, we want this to be dead simple for users. You shouldn't need a PhD in CUDA to run a GPU job. Just tell us you
need a GPU with 8GB of memory, and we'll handle the rest. All that complexity with driver versions, device nodes,
library paths? That's our problem, not yours.

But we're not compromising on security. Every GPU access goes through the same strict controls as CPU and memory. By
default, jobs get zero GPU access - they have to explicitly ask for it, and we have to explicitly grant it. No
exceptions.

We're also being smart about resource usage. GPUs are expensive. Sometimes you need a whole GPU to yourself (like when
training a big model), but sometimes you're just running inference and barely using 10% of it. We'll support both
modes - exclusive when you need it, shared when you don't.

And yeah, CUDA versions are a pain. One job needs CUDA 11.8, another needs 12.2. We'll handle multiple versions and make
sure each job gets what it needs without conflicts

### The Main Pieces

#### GPU Manager

This is the brain of the operation. It's a new component that keeps track of all the GPUs in the system. When Joblet
starts up, it pokes around in `/proc/driver/nvidia/gpus/` to see what GPUs you have. It figures out how much memory they
have, what compute capability they support, whether they're healthy, all that stuff.

The GPU Manager is also the gatekeeper. When a job wants a GPU, it goes through the manager. The manager decides which
GPU to give it based on what's available and what strategy you've configured. Maybe you want to pack jobs onto GPUs to
save power, or maybe you want to spread them out to avoid thermal throttling. Your choice.

#### Making Isolation Work with GPUs

This was the hard part. We already have this great isolation system using pure Linux kernel features - no Docker, no
containerd, just namespaces and cgroups. GPUs weren't designed with this in mind, but that's what makes it interesting.

We're using the cgroups v2 device controller directly - the same kernel feature that Docker uses under the hood, but
without all the container baggage. Each job only sees the GPU devices we explicitly allow through device cgroup rules.
We create the necessary device nodes (`/dev/nvidia0`, `/dev/nvidiactl`, etc.) inside the job's isolated filesystem using
mknod, but only for the GPUs that job is allowed to use.

For CUDA libraries, we're using bind mounts - again, native Linux kernel feature, no container runtime needed. We mount
them read-only into the job's mount namespace. This way jobs can't mess with the libraries, but they can use them. We
also set up environment variables to enforce memory limits - it's not perfect, but it works well enough and doesn't
require any external tools.

#### API Changes

We kept the API changes minimal and clean. When you submit a job, you can now include GPU requirements - how many GPUs,
how much memory, what compute capability. The status responses now include GPU allocation details and metrics so you can
see what's actually happening with your GPUs.

#### Smarter Scheduling

The scheduler got a lot smarter about GPUs. It understands that some workloads don't play nice together - like don't put
a training job next to an inference job on the same GPU because they have different memory access patterns. It can also
spread jobs across GPUs to balance thermal load, which actually matters when you're running these things at full tilt

## Consequences

### What We're Excited About

This is going to be really good. We're basically getting enterprise-grade GPU support that rivals what you'd get with
Kubernetes or Slurm, but staying true to Joblet's philosophy - no containers, no orchestrators, just native Linux kernel
features. You still get the simple, direct Joblet experience, just now with GPUs.

The security story is solid. We're not cutting any corners here - GPU memory stays isolated between jobs using the same
kernel-level isolation we use for everything else. The cgroups device controller (the exact same kernel feature Docker
uses, but we're using it directly) means a job can't just grab any GPU it wants. It's as locked down as our CPU and
memory isolation.

The best part? You don't need to be a GPU expert to use this. The system figures out what drivers you have, what CUDA
versions are available, and matches everything up automatically. Users just say "I need 2 GPUs with 16GB each" and we
make it happen.

We've also built in flexibility from day one. You can optimize for different things - pack jobs tightly to save power,
spread them out to avoid heat issues, or give certain jobs exclusive GPU access when they really need it. And when
NVIDIA's MIG technology takes off, we're ready for it. Same with AMD GPUs when we get to them.

### What's Going to Be Harder

Let's be honest - this is adding a lot of complexity. GPUs are messy. Different driver versions behave differently, CUDA
has its own quirks, and every GPU generation has its own special features. Our codebase is going to get bigger and
harder to maintain.

We're also betting heavily on NVIDIA for now. Yes, we designed it to support other vendors later, but actually adding
AMD or Intel support is going to be real work, not just a configuration change.

There's overhead too. We need to poll nvidia-smi to get metrics, track GPU state, manage allocations - all that takes
CPU cycles. We think we can keep it under 2% overhead, but it's not free.

Testing is going to be interesting. You can't really test GPU code without GPUs, and cloud CI/CD systems with GPUs are
expensive. We'll need to build good mocks and be really careful with our abstractions.

### How We're Rolling This Out

We're not doing this all at once. That would be crazy. Here's how we're thinking about it:

**Phase 1 - The Basics**: Get GPU detection working. Figure out what GPUs are in the system, track their state, build
the basic allocation logic. Nothing fancy, just "here's a GPU, here's a job that wants it."

**Phase 2 - Integration**: Wire it into our isolation system. This is where it gets interesting - making cgroups play
nice with GPU devices, getting device nodes to show up in the right places, making sure jobs can actually use the GPUs
we give them.

**Phase 3 - The CUDA Maze**: Sort out the library situation. Detect what CUDA versions are installed, figure out how to
mount them into jobs, handle version mismatches gracefully. This is probably the most annoying part.

**Phase 4 - Advanced Features**: Now we can have some fun. Memory limits, maybe MIG support if the hardware supports it,
smarter scheduling algorithms. The stuff that makes it actually nice to use.

**Phase 5 - Production Ready**: Testing, testing, and more testing. Performance tuning. Writing documentation that
actually helps people. Making sure we didn't break anything for non-GPU users.

### For Existing Users

If you're already running Joblet, don't worry - nothing changes unless you want it to. GPU support is off by default.
When you're ready to try it, you flip a switch in the config. If something goes wrong, flip it back. Your non-GPU jobs
keep running the whole time.

### Security Stuff We're Worried About

GPUs weren't designed with multi-tenancy in mind, so we need to be paranoid:

- We're clearing GPU memory between jobs. Yes, it takes time, but data leakage is not acceptable.
- We're putting limits on how long GPU kernels can run. No mining crypto on someone else's GPU.
- We're rate-limiting GPU allocations so you can't DOS the system by requesting and releasing GPUs in a tight loop.
- Everything gets logged. Who used what GPU, when, for how long, how much memory - it's all there for the auditors

## What Else We Looked At

Before settling on this design, we kicked around a few other ideas:

**Just use nvidia-container-toolkit**: This is what Docker and containerd use for GPU support. But here's the thing -
Joblet's whole philosophy is to use native Linux features directly, not through container runtimes. Adding
nvidia-container-toolkit would mean bringing in Docker or containerd as a dependency, which goes against everything
Joblet stands for. We built this system to prove you don't need containers for isolation - just the kernel features that
containers themselves use.

**Wrap nvidia-docker**: This is even further from our philosophy. Now we're not just using container runtimes, we're
specifically tying ourselves to Docker. Joblet exists because we believe in using Linux kernel capabilities directly -
namespaces, cgroups, bind mounts - without the overhead and complexity of container systems. Adding Docker just for GPU
support would be admitting defeat.

**YOLO Device Passthrough**: The simplest option - just pass through the GPU device nodes and call it a day. This
actually aligns with our philosophy of simplicity, but it fails on the security and multi-tenancy front. No resource
limits, no proper isolation, jobs stepping on each other's toes. We can do better using native kernel features.

**Build on top of NVIDIA's Kubernetes Device Plugin**: We looked at this pretty hard. They've solved similar problems,
but their solution requires the entire Kubernetes stack. Again, this goes against Joblet's core principle - we're
proving you can do sophisticated resource management and isolation using just Linux kernel features, no orchestration
platform required.

The approach we're going with stays true to Joblet's philosophy: use native Linux kernel features directly. We're using
cgroups v2 for device control, mount namespaces for library isolation, and the /proc filesystem for device discovery. No
containers, no orchestrators, just the kernel doing what it does best. Yeah, it's more work to build from scratch, but
that's the whole point of Joblet - showing that you can build powerful isolation and resource management without all the
traditional containerization overhead.

## Things We're Reading

If you want to dig deeper into how this stuff works:

- The [NVIDIA driver documentation](https://docs.nvidia.com/cuda/cuda-installation-guide-linux/) is actually pretty good
  once you get past all the marketing speak
- The [Kubernetes GPU device plugin code](https://github.com/NVIDIA/k8s-device-plugin) showed us a lot of edge cases we
  hadn't thought about
- The [cgroups v2 documentation](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html) is dry but essential
  for understanding device control
- NVIDIA's [MIG user guide](https://docs.nvidia.com/datacenter/tesla/mig-user-guide/) for when we get to the fancy GPU
  partitioning stuff
- The [CUDA compatibility docs](https://docs.nvidia.com/deploy/cuda-compatibility/) because version mismatches will make
  you cry