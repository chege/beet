File-set Design Proposals for AI Coding CLI

Below are five distinct file-set designs for structuring unstructured requirements into a repository of AI-readable
instructions. Each proposal outlines the files included, their roles, core/optional status, and how they meet goals like
safety, progress tracking, and resumability.

Proposal 1: Minimalist Blueprint (Simplicity-Focused)
• Files and Roles:
• AGENTS.md (core) – Entry-point instructions for the coding agent, including a brief summary of the goal and references
to other files. It provides high-level guidance and ensures the agent starts with a clear objective and knowledge of
available instructions.
• RULES.md (core) – A compact list of coding standards, style guidelines, and “do’s and don’ts” specific to the project.
This makes coding rules explicit (e.g. naming conventions, dependency rules) without cluttering the main instruction ￼.
• TASKS.md (core) – A simple ordered checklist of implementation steps (each a small work unit). Tasks are listed as
bullet points or checkboxes that the agent can check off as it completes them. This file breaks the project into clear,
incremental tasks.
• Optimization Focus: This design optimizes for simplicity and minimal overhead. With only a few files, it’s easy for
the agent to parse the structure. All essential information (goal, rules, tasks) is present without duplication,
reducing the chance of confusion.
• Progress Tracking: Progress is represented directly in TASKS.md via checklist markers. As the agent completes a task,
it can mark it (e.g. [x] done). The remaining unchecked items show outstanding work. This simple markdown checklist
approach lets the agent persist and recall completed vs. pending tasks ￼.
• Rationale: Separating into these three files keeps concerns distinct yet minimal. AGENTS.md gives a single starting
context, RULES.md ensures coding constraints are first-class, and TASKS.md focuses on execution steps. This is useful
for the CLI because it provides just enough structure to guide an AI agent (which reads the repo files) without
overwhelming it. The agent can easily find the project rules and the to-do list, enabling safe, step-by-step development
in a reproducible way.

Proposal 2: Comprehensive Separation (Clarity & Traceability)
• Files and Roles:
• AGENTS.md (core) – Master instruction file orchestrating the process. It outlines the overall intent and lists all
auxiliary files. Serves as the guided entrypoint so the agent understands the project structure from the start ￼.
• INTENT.md (core) – A structured interpretation of the original requirements. It restates the free-form input as clear
objectives, scope, and acceptance criteria. This file ensures the agent knows exactly what needs to be built or changed.
• DESIGN.md (optional) – High-level solution outline or technical design decisions (if the task warrants design
planning). It might describe the approach, key modules to modify, and any architectural notes. This is included for
complex features where planning the solution is beneficial before coding.
• RULES.md (core) – Comprehensive coding and quality guidelines. This expands on coding style, architectural
constraints, performance or security requirements, and any testing standards. By isolating rules here, the agent has a
single source of truth for all quality expectations.
• PLAN.md (core) – A detailed task breakdown. This lists all work units (possibly grouped into phases or components)
that the agent must execute. Tasks may be numbered or nested to reflect sub-tasks. The plan is thorough, covering
implementation and test updates for each item.
• PROGRESS.md (optional) – An explicit progress tracker. It could mirror the tasks from PLAN.md with markers or
timestamps for completion. For example, each task entry might be updated with a status (pending/completed). This file is
generated or updated as work proceeds, acting as a log or checklist separate from the original plan.
• Optimization Focus: This file set maximizes clarity and traceability. Each aspect of the project is in its own
document, which makes it easier to audit and review the AI’s context. The agent is less likely to mix up concerns since
requirements, design, rules, and task plan are clearly delineated. This design is optimized for safety and scale – it
can handle complex projects by breaking information into digestible parts and provides a paper trail for each decision.
• Progress Tracking: Progress is visible through the PROGRESS.md log or by the state of PLAN.md tasks. Each task’s
status is recorded (e.g. checked off or moved to PROGRESS.md when done). By comparing the plan with the progress log,
one can infer which work units are completed and which remain. This separate progress artifact makes it easy to resume
work after a pause – the agent can read what’s completed in PROGRESS.md without relying on commit history alone.
• Rationale: This separation of files is useful for the CLI’s purpose because it reduces ambiguity by isolating
concerns. The agent reads a well-structured repository: the intent file to grasp “what and why”, the design (if present)
to understand “how”, the rules to know constraints, and the plan to know “what to do next”. This structure guides
automated code changes in a controlled, reviewable way. It also provides complete transparency – anyone (or any tool)
inspecting the repo can see the original intent and how the work was planned and executed step by step. The presence of
a dedicated progress file further ensures that work can stop and restart cleanly, with full knowledge of the state.

Proposal 3: Progress-Driven Workflow (Resumability-Focused)
• Files and Roles:
• AGENTS.md (core) – The primary control document directing the agent. It contains the mission overview and references
the task workflow. It might instruct the agent to iterate through the tasks one by one and how to update progress.
• OBJECTIVES.md (core) – A brief outline of the overall goal and sub-goals. This is similar to a requirements summary or
“definition of done”. It ensures the agent always keeps the end-goal in mind as it works through the tasks.
• TODO.md (core) – A living task list that doubles as a progress tracker. Tasks are listed with checkboxes or status
tags (e.g. “TODO”, “DONE”). The agent updates this file as each work unit is completed. This single source shows both
the to-do items and their completion state in-line.
• CONTEXT.md (optional) – Relevant context or pointers for the tasks. For example, it may list key files or modules
related to each task, or the current state of the codebase that the agent should be aware of. This prevents the agent
from spending time re-discovering context and helps it resume after breaks with full knowledge of where it left off.
• Optimization Focus: This design optimizes for resumability and iterative progress. It is all about making sure the
agent can pause and resume work methodically. The structure is lean but strongly centered on the TODO.md which always
reflects the current state. It’s also focused on helping the agent maintain context over time, so it doesn’t forget what
it was doing mid-way ￼ ￼.
• Progress Tracking: Progress is front-and-center in TODO.md. Each task entry serves as a progress marker (unchecked
means pending, checked means done). The agent (or a supervising system) can infer progress at a glance by scanning this
file. Because the agent updates the same file as it works, there’s a persistent record of completion. Even after a long
pause, the next session can read TODO.md to see which tasks remain, ensuring continuity.
• Rationale: Separating files in this way is useful because it turns the workflow into a transparent checklist that the
agent follows. AGENTS.md provides the high-level guidance, but the agent’s day-to-day operations are driven by TODO.md.
This keeps the agent focused on one item at a time, which improves reliability. The OBJECTIVES.md and optional
CONTEXT.md further make sure the agent doesn’t lose sight of the bigger picture or important details while grinding
through tasks. Overall, this structure makes the execution deterministic (the agent follows a clear sequence) and makes
the state of the work obvious to any tool or future run that inspects the repo.

Proposal 4: Quality-Gated Pipeline (Safety & Testing-Focused)
• Files and Roles:
• AGENTS.md (core) – The coordinator file that instructs the agent on the workflow with an emphasis on quality. It
outlines the feature goal and explicitly tells the agent to enforce quality checks (like running tests) at each step.
• GUIDELINES.md (core) – A detailed set of coding standards, best practices, and project policies. This includes style
rules, security guidelines, performance considerations, and any regulatory or dependency rules. The agent uses this as a
rulebook to ensure all changes meet the pre-defined quality bar. (For example: “All new functions must have unit tests”,
“Avoid global variables”, etc.)
• TEST_PLAN.md (core) – A file dedicated to testing requirements. It lists the test cases to create or update in tandem
with the implementation tasks. For each major functionality change, this plan describes how it will be verified (unit
tests, integration tests, etc.). This ensures the agent knows testing is not optional but part of the definition of
done.
• STEP_BY_STEP.md (optional) – A sequential list of implementation steps gated by checks. Each step in this list
corresponds to a small code change or commit. Importantly, after each step, the agent is instructed to run validations (
like linting and tests) before proceeding to the next. The file acts like a checklist with quality gates; e.g., “Step 3:
Implement feature X (run tests and linter – must pass)”.
• Optimization Focus: This design optimizes for safety and code quality. By front-loading strict guidelines and a test
plan, the agent is constrained to high-quality outputs. It also encourages small, incremental commits (through the
step-by-step list), which makes the process reproducible and reviewable. The emphasis is on not just finishing tasks,
but finishing them correctly and verifying each one.
• Progress Tracking: Progress is reflected in the step-by-step checklist and the test plan outcomes. Each
STEP_BY_STEP.md entry can be marked complete once its code change is made and all corresponding tests from TEST_PLAN.md
pass. This way, “completed” truly means done to quality standards. The agent can update the steps file with a checkmark
when a step is finished and verified. By checking this file, one can infer which steps are done (and were validated) and
which are still pending or failed. This structured progression ensures that work can resume safely: any paused state
will have clearly which step was last completed and that all prior steps met the acceptance criteria.
• Rationale: Separating files by quality focus is very useful in this CLI’s context because it embeds quality control
into the development workflow. GUIDELINES.md acts as an always-present safety net, reminding the agent of general coding
rules (preventing common errors or undesirable patterns ￼). TEST_PLAN.md makes test creation first-class, so the agent
treats tests as part of the requirements, not an afterthought. The STEP_BY_STEP.md sequence enforces incremental
development – the agent is effectively guided to “build, then verify” repeatedly. This separation means the automation
won’t easily skip tests or ignore standards; each file reinforces the expectation that code and tests move forward
together. For the CLI user, this yields more trustworthy code changes that are consistently verified and easier to
review (small, verified increments).

Proposal 5: Modular Phase Breakdown (Scalability-Focused)
• Files and Roles:
• AGENTS.md (core) – The high-level orchestration file explaining the multi-phase or modular approach. It tells the
agent how the work is divided (e.g., by component or phase) and in what order to tackle them. This ensures the agent
knows there are multiple segments to the project and references each segment’s instructions.
• REQUIREMENTS.md (core) – A full breakdown of the feature requirements partitioned by modules or subsystems. For
example, it might be organized into sections like Frontend, Backend, Database, etc., each with specific requirements.
This helps the agent see the big picture and the parts, which is crucial for larger-scale tasks.
• CONTEXT.md (core) – A repository or system overview. This file provides background on the existing project structure,
key modules, and relationships. For instance, it might outline “Module A handles X, Module B handles Y” or list
important files to be aware of. By giving this context, the agent can navigate a large codebase more efficiently and
avoid expensive exploratory steps ￼.
• PLAN_main.md (core) – A top-level plan listing major work units, possibly aligned with the modules/phases from the
requirements. It enumerates each big task group (e.g., “Implement Frontend UI changes”, “Update API endpoints”, “Migrate
database schema”) and may reference sub-plans for details. This file acts as a roadmap for the agent.
• Module-specific plan files (optional) – For each major module or phase, there can be a separate tasks file (e.g.,
PLAN_frontend.md, PLAN_backend.md). These contain detailed tasks for that segment only. They are optional and generated
only if a project is complex enough to warrant isolating tasks by area. Each such file focuses on that component’s
changes and testing needs.
• Optimization Focus: This design is optimized for handling complex or large-scale projects. It emphasizes modularity
and parallel clarity. By splitting instructions by domain or phase, the agent can focus on one section at a time with
full context, which helps manage complexity and context window limits. It also improves maintainability – if a
particular module’s plan changes, it’s localized to its own file.
• Progress Tracking: Progress can be inferred at both high and granular levels. The top-level PLAN_main.md can indicate
which major groups are done (e.g., by marking a whole section as completed once its sub-tasks are finished). Within each
module-specific plan, tasks are checked off as completed. This two-tier tracking means one can quickly see overall
progress (which phases are finished) and drill down into detailed progress within each phase. After a pause, the agent
can resume work by looking at PLAN_main.md to find the next unfinished area, then proceed to the corresponding detailed
plan file. All progress is persistent in these markdown plans.
• Rationale: This separation is very useful for the CLI’s use case because it mirrors a structured development approach
that scales. The unstructured input is transformed into an organized project blueprint. By dividing by modules, the
agent is less likely to be overwhelmed or to inadvertently mix concerns (for example, front-end vs back-end logic). Each
file serves as a single source of truth for a part of the project – requirements clarify what to do, context gives the
lay of the land, and plans give exact steps. This is especially beneficial when the project spans multiple domains: the
agent can independently handle each domain’s tasks while still being guided by the overarching plan in AGENTS.md and
PLAN_main.md. The result is deterministic, stepwise execution that makes the agent’s work in a big project safe,
discoverable, and easy to pause and resume without losing track.