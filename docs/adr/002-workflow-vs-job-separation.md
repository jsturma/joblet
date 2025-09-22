# ADR-002: Workflow vs Individual Job Architecture

## Status

Accepted

## Context

Early in Joblet's development, we had a decision to make about workflows. The easy path would have been to treat
workflows as just a special type of job - maybe a job that spawns other jobs. Many systems take this approach, and it
works... until it doesn't.

The thing is, workflows and individual jobs have fundamentally different concerns. A job wants to run a command, capture
output, manage resources, and report status. Pretty straightforward. A workflow, on the other hand, needs to think about
dependencies, orchestration, failure handling across multiple jobs, and complex state management. Trying to shoehorn
both into the same abstraction felt like we were heading for a world of pain.

We also looked at how users think about these concepts. When someone runs a single job, they want immediate execution
and simple status. When they define a workflow, they're thinking about pipelines, data flow, and coordination. These are
different mental models that deserve different implementations.

## Decision

We decided to implement completely separate service layers for jobs and workflows. Jobs get their own `JobServiceServer`
with methods focused on direct execution. Workflows get a `WorkflowServiceServer` that handles orchestration and
dependency management.

The key insight was that jobs don't need to know about workflows at all. A job just runs. If that job happens to be part
of a workflow, the workflow service handles that relationship. The job itself remains blissfully unaware.

This separation extends through the entire stack. Different protobuf services, different command hierarchies in the
CLI (`rnx job` vs `rnx workflow`), different internal managers, different storage patterns.

## Consequences

### The Good

The separation has been fantastic for code clarity. When you're working on job execution, you're not wading through
workflow orchestration logic. When you're debugging a workflow issue, you're not distracted by low-level process
management code.

It's also made the system more robust. Issues in the workflow orchestrator don't affect simple job execution. We can
optimize each path independently - jobs for low latency, workflows for complex coordination.

The mental model for developers (both us and users of the API) is crystal clear. You know exactly which service to call
for what you need. The API surface is intuitive because it matches how people think about the problem.

### The Trade-offs

There's definitely some code duplication. Both services need to track status, both need to handle resource limits, both
need to manage storage. We've addressed this through shared libraries and interfaces, but it's still more code than a
unified approach.

The client needs to understand which service to call. This adds a small amount of complexity to the CLI, though we've
hidden it well behind the command structure.

### The Unexpected Benefits

What surprised us was how this architecture made testing easier. We can test job execution in complete isolation without
any workflow machinery. We can test workflow orchestration with mock jobs. The test suites are focused and fast.

It also opened up optimization opportunities we didn't expect. Since workflows typically run longer and have different
performance characteristics than individual jobs, we can tune their resource management differently. Workflows get more
aggressive caching, different timeout strategies, and specialized monitoring.

The separation also made it trivial to add workflow-specific features like visualization, progress tracking, and
dependency analysis without cluttering the job execution path.

## Learn More

See [workflow.md](/docs/recommendations/workflow.md) for detailed workflow implementation
and [unified-architecture-implementation.md](/docs/recommendations/unified-architecture-implementation.md) for how the
pieces fit together.