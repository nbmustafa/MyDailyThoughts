The **Claude Certified Architect – Foundations (CCA-F)** is a real, live exam launched by Anthropic on March 12, 2026. It's a 120-minute, 60-question proctored exam covering five domains: Agentic Architecture (27%), Claude Code (20%), Prompt Engineering (20%), Tool Design & MCP (18%), and Context Management (15%). [LinkedIn](https://www.linkedin.com/posts/analytics-india-magazine_anthropicai-claudeai-aicertification-activity-7439256691344351232-XzCa)

The actual exam questions are confidential and proctored. Below is a full set of **60 high-quality practice questions** modeled on the official exam guide's domains, task statements, scenarios, and anti-patterns — the closest thing to the real exam you can study with.

---

## CCA-F: 60 Practice Questions

---

### 🔷 DOMAIN 1 — Agentic Architecture & Orchestration (16 Questions)

**Q1.** A customer support agent using the Claude Agent SDK is running an agentic loop. After calling a `lookup_order` tool, Claude returns a message with `stop_reason: "tool_use"`. What is the correct next action?

- A) Present Claude's response to the user and wait for input
- B) Execute the requested tool and append the result to the conversation, then send the next API request
- C) Reset the session and retry the query
- D) Treat the loop as complete and return the partial response

✅ **B** — The loop continues: execute the tool, append its result, and send back to Claude.

---

**Q2.** You are building a multi-agent research pipeline. The coordinator agent spawns three subagents to research different topics, then synthesizes their outputs. Which pattern best describes this architecture?

- A) Pipeline pattern
- B) Hub-and-spoke orchestration
- C) Fan-out / fan-in pattern
- D) Sequential chain-of-thought

✅ **C** — Fan-out (parallel subagents) then fan-in (synthesis) is the correct pattern name in this context.

---

**Q3.** A production agentic loop is handling financial transactions. You need to enforce that no refund over $500 is processed without human approval. What is the MOST reliable mechanism?

- A) Add "Do not process refunds over $500 without approval" to the system prompt
- B) Use a PostToolUse hook to programmatically intercept calls to `process_refund` and trigger escalation when the amount exceeds $500
- C) Set `max_tokens` to limit Claude's ability to generate large refund amounts
- D) Train a custom classifier to detect high-value refund intents

✅ **B** — PostToolUse hooks are the correct, reliable mechanism for programmatic guardrails. System prompts alone are not enforced at the code level.

---

**Q4.** Claude returns `stop_reason: "end_turn"` during an agentic loop. What should the agent do?

- A) Continue looping and make another API call
- B) Spawn a subagent to handle the remaining task
- C) Exit the loop and return the response to the user
- D) Retry the last tool call

✅ **C** — `end_turn` signals Claude has completed its response. The loop should exit.

---

**Q5.** You have a multi-agent system where a coordinator delegates document summarization to a subagent. The subagent completes its task but the coordinator's context window is approaching its limit. What is the BEST strategy?

- A) Increase `max_tokens` to handle the additional content
- B) Have the subagent return a structured summary with only key extracted facts, not the full output
- C) Restart the coordinator's session with all previous context re-injected
- D) Switch to a smaller model for the coordinator

✅ **B** — Subagents should return distilled, structured outputs to prevent coordinator context overload.

---

**Q6.** An agentic loop has been running for 20 turns processing a large task. You notice Claude is starting to lose track of earlier context. What is the recommended architectural solution?

- A) Increase the model's context window by switching to a larger model tier
- B) Use progressive summarization to condense prior conversation turns into a structured summary block
- C) Reset the session after every 10 turns
- D) Disable tool calls to reduce token usage

✅ **B** — Progressive summarization is the production pattern for long agentic sessions.

---

**Q7.** Which of the following is a valid trigger for escalating an agentic task to a human reviewer?

- A) The agent's self-assessed confidence score drops below 70%
- B) Sentiment analysis of the user's message indicates frustration
- C) A policy gap exists that the agent is not authorized to resolve
- D) The agent has made more than 5 tool calls in a single session

✅ **C** — Policy gaps, explicit user requests for human assistance, and capability limits are the three valid escalation triggers. Agent self-assessment and sentiment are unreliable.

---

**Q8.** You want to prevent a subagent from calling tools outside its designated role. What is the BEST approach?

- A) Include a warning in the system prompt about unauthorized tool use
- B) Provide the subagent only with the tools relevant to its specific role when initializing the agent
- C) Monitor tool calls post-execution and retroactively reject unauthorized ones
- D) Use a single agent for all roles to simplify authorization

✅ **B** — Restricting the tool set at initialization is the correct architectural pattern.

---

**Q9.** In the Claude Agent SDK, what is the purpose of a PreToolUse hook?

- A) To log tool outputs for debugging after a tool runs
- B) To normalize or block a tool call before it is executed
- C) To retry a failed tool call automatically
- D) To summarize the tool result before returning it to Claude

✅ **B** — PreToolUse hooks intercept the tool call before execution, enabling validation, blocking, or transformation.

---

**Q10.** A coordinator agent needs to ensure the extract_metadata tool is always called before any enrichment tools in a workflow. What is the correct mechanism?

- A) Describe the ordering requirement in the system prompt
- B) Use `tool_choice` with forced selection to require `extract_metadata` first, then handle subsequent steps in follow-up turns
- C) Hard-code a sequential tool execution schedule in the application layer
- D) Use a PostToolUse hook on enrichment tools to verify metadata was extracted

✅ **B** — `tool_choice` forced selection is the Claude-native mechanism for enforcing tool call ordering.

---

**Q11.** What does `tool_choice: "any"` guarantee in a Claude API request?

- A) Claude will select the most appropriate tool based on the prompt
- B) Claude must call at least one tool and cannot return a conversational text-only response
- C) Claude will call all available tools in sequence
- D) Claude can choose between a tool call or a text response

✅ **B** — `"any"` forces Claude to call a tool rather than generating a plain text reply.

---

**Q12.** A multi-agent research system uses three parallel subagents to gather information and a synthesis agent to combine results. The synthesis agent occasionally hallucinates facts not present in the subagent outputs. What is the BEST architectural fix?

- A) Increase the synthesis agent's temperature parameter
- B) Add a verify_fact tool scoped to the synthesis agent that cross-references its output against subagent results
- C) Reduce the number of subagents to minimize conflicting inputs
- D) Add "Do not hallucinate" to the synthesis agent's system prompt

✅ **B** — A scoped verify_fact tool gives the synthesis agent a mechanism to ground its outputs.

---

**Q13.** Session state management in the Claude Agent SDK is BEST handled by:

- A) Storing session state in Claude's context window across all turns
- B) Persisting session state externally and injecting the relevant state at the start of each session
- C) Using cookies in the API headers to maintain session continuity
- D) Relying on Claude's implicit memory of previous interactions

✅ **B** — Sessions are stateless; external persistence with selective injection is the production pattern.

---

**Q14.** An agentic loop has been calling a tool that intermittently times out. The agent should retry transient failures but not business-logic errors. How should error handling be structured?

- A) Retry all errors up to 3 times before escalating
- B) Catch all exceptions and present a generic error message to the user
- C) Classify errors by type: retry transient errors (network timeouts), escalate business errors (invalid input, authorization failures), and propagate fatal errors
- D) Use a try/catch block in the system prompt instructions

✅ **C** — Structured error classification drives appropriate agent behavior.

---

**Q15.** You are designing a coordinator-subagent system. The coordinator needs to pass a customer's account ID and order history to a subagent for processing. What is the BEST method?

- A) Ask the user to re-enter their account information when the subagent begins
- B) Include the relevant context in the subagent's initial system or user message when spawning it
- C) Store the data in Claude's context window and expect the subagent to retrieve it automatically
- D) Use a shared global variable accessible to all agents

✅ **B** — Context must be explicitly passed when spawning subagents; agents have no implicit shared state.

---

**Q16.** What is the key distinction between model-driven decision-making in an agentic system and a pre-configured decision tree?

- A) Model-driven systems are faster and cheaper than decision trees
- B) In model-driven systems, Claude reasons about which tool to call next based on conversation context; decision trees follow pre-configured sequences regardless of context
- C) Decision trees use Claude's reasoning while model-driven systems rely on rule-based logic
- D) There is no meaningful distinction; both approaches produce equivalent results

✅ **B** — This distinction is fundamental to agentic architecture design.

---

### 🔷 DOMAIN 2 — Tool Design & MCP Integration (11 Questions)

**Q17.** An MCP tool description reads: `"tool": "db_query"`. During testing, Claude frequently bypasses this tool in favor of built-in Grep for data lookups. What is the most likely root cause?

- A) The tool is not registered in the MCP server correctly
- B) The tool description is too short and does not explain the tool's capabilities or output, so Claude cannot distinguish it from built-in alternatives
- C) Claude always prefers built-in tools over MCP tools by design
- D) The tool's response time is too slow

✅ **B** — Rich, detailed tool descriptions are the primary driver of tool selection. Vague names alone are insufficient.

---

**Q18.** Where should a shared MCP server configuration be stored so it is version-controlled and shared across a development team?

- A) `~/.claude.json` (user-level config)
- B) `.mcp.json` in the project root (project-level config, committed to VCS)
- C) In environment variables only
- D) In the system prompt of each Claude Code session

✅ **B** — `.mcp.json` at project scope is tracked in source control; `~/.claude.json` is user-level/personal.

---

**Q19.** An MCP tool for processing refunds returns the following error: `{ "isError": true, "content": "Error" }`. Why is this response problematic for an agentic system?

- A) The `isError` flag should be set to `false` for non-fatal errors
- B) The generic error message gives Claude no actionable information — it cannot determine whether to retry, change the query, or escalate
- C) MCP tools should never return errors; they should return empty objects instead
- D) The error is formatted correctly and will cause no issues

✅ **B** — Structured, descriptive error messages are essential for agent decision-making.

---

**Q20.** You are building an MCP server for a document management system. You want agents to know what documents are available without requiring exploratory tool calls. What MCP feature enables this?

- A) Tool descriptions with file path examples
- B) MCP Resources, which provide agents with a browseable content catalog
- C) A `list_documents` tool that returns all available documents
- D) Environment variable injection at server startup

✅ **B** — MCP Resources expose content catalogs so agents understand available data without probing.

---

**Q21.** A team is integrating a third-party project management service (Jira) into their Claude workflow. Should they build a custom MCP server or use an existing community MCP server?

- A) Always build custom; community servers cannot be trusted
- B) Use an existing community MCP server for standard integrations like Jira; reserve custom servers for team-specific workflows not covered by community solutions
- C) Build custom to ensure full feature parity
- D) Use direct HTTP API calls instead of MCP for all third-party services

✅ **B** — Community MCP servers should be preferred for standard integrations; custom servers for proprietary workflows.

---

**Q22.** What is the BEST practice when deciding whether to split one large MCP tool into multiple smaller tools?

- A) Always keep all functionality in a single tool to minimize API calls
- B) Split when a tool has too many optional parameters and agents frequently misroute calls; consolidate when tools are always called together
- C) Split only when the tool exceeds 1,000 characters in its description
- D) Always split tools into the smallest possible atomic units

✅ **B** — Tool granularity decisions should be driven by agent routing accuracy and usage patterns.

---

**Q23.** Which structured error response type should an MCP tool return when a user lacks permission to perform an action?

- A) Transient error (retry recommended)
- B) Fatal error (terminate session)
- C) Permission error (do not retry; escalate or inform user)
- D) Business logic error (retry with corrected inputs)

✅ **C** — Permission errors are non-retryable and require escalation or user notification.

---

**Q24.** An MCP server requires an authentication token. What is the BEST practice for managing this token in a team development environment?

- A) Hard-code the token in the `.mcp.json` file
- B) Use environment variable expansion in `.mcp.json` to reference a variable like `$JIRA_API_TOKEN`, set locally by each developer
- C) Store the token in the Claude Code session's context window
- D) Prompt Claude to request the token from the user at the start of each session

✅ **B** — Environment variable expansion keeps secrets out of version control.

---

**Q25.** You want Claude to always call a specific tool before proceeding with other analysis. Which `tool_choice` configuration enforces this?

- A) `tool_choice: "auto"`
- B) `tool_choice: { "type": "tool", "name": "required_tool_name" }`
- C) `tool_choice: "any"`
- D) `tool_choice: "none"`

✅ **B** — Specifying the tool name in `tool_choice` forces Claude to call that exact tool.

---

**Q26.** An MCP tool for data extraction returns verbose JSON with hundreds of fields. Agents downstream are experiencing context overload. What is the BEST solution?

- A) Increase the model's context window
- B) Redesign the tool to return only the fields the agent needs, or add a `fields` parameter to allow selective field retrieval
- C) Instruct Claude to ignore irrelevant fields in the system prompt
- D) Split the JSON output across multiple tool calls

✅ **B** — Tool output should be scoped to what downstream agents actually need.

---

**Q27.** When should you use `isError: true` versus returning an error as structured content in an MCP tool response?

- A) Use `isError: true` for all error types; never embed errors in content
- B) Use `isError: true` to signal an execution failure so the agent can branch appropriately; use structured error content within the response to provide actionable details (error type, suggested action)
- C) Use structured content only; `isError` is deprecated in newer MCP versions
- D) Use `isError: true` only for network errors; use content for business errors

✅ **B** — Both are used together: the flag signals failure, while structured content guides the agent's next action.

---

### 🔷 DOMAIN 3 — Claude Code Configuration & Workflows (12 Questions)

**Q28.** You have a monorepo with a frontend and backend directory. The backend has strict type safety rules that don't apply to the frontend. How should you configure Claude Code?

- A) Add all rules to the root `CLAUDE.md` and use comments to indicate which apply where
- B) Create a root `CLAUDE.md` for shared rules and separate `CLAUDE.md` files in `/frontend` and `/backend` for path-specific rules
- C) Use a single flat `CLAUDE.md` and override rules in each session via slash commands
- D) Configure rules only at the user level in `~/.claude.json`

✅ **B** — CLAUDE.md files cascade hierarchically; directory-level files override/extend the root for their scope.

---

**Q29.** What flag must be passed when running Claude Code in a CI/CD pipeline to prevent it from entering interactive mode and waiting for user input?

- A) `--batch`
- B) `--non-interactive`
- C) `-p` (or `--print`)
- D) `--silent`

✅ **C** — The `-p` / `--print` flag enables non-interactive, single-turn mode for CI/CD integration.

---

**Q30.** A team wants to automate PR reviews with Claude Code. They want structured JSON output containing issue severity ratings. Which flags should be combined?

- A) `--output text --format json`
- B) `-p --output-format json --json-schema <schema_file>`
- C) `--batch --structured`
- D) `--interactive --json`

✅ **B** — `-p` for non-interactive mode combined with `--output-format json` and `--json-schema` produces validated structured output.

---

**Q31.** What does the `context: fork` option in a Claude Code Skill's frontmatter do?

- A) Creates a backup copy of the current session state
- B) Runs the Skill in an isolated subagent context, preventing its verbose output from polluting the main session's context window
- C) Forks a new Git branch for the Skill's code changes
- D) Splits the skill execution across two parallel Claude instances

✅ **B** — `context: fork` is specifically used to isolate Skills that produce large outputs.

---

**Q32.** A developer wants Claude Code to automatically run tests and lint checks after every file edit. Where is this configured?

- A) In the system prompt of each Claude Code session
- B) Using PostToolUse hooks in the Claude Code configuration (settings.json or hooks config) to trigger test/lint commands after Write or Edit tool calls
- C) By adding a custom slash command `/post-edit-checks`
- D) By specifying a `run_after_edit` key in `CLAUDE.md`

✅ **B** — PostToolUse hooks wired to Write/Edit tools are the correct mechanism for automated post-edit workflows.

---

**Q33.** What is Plan Mode in Claude Code and when should it be used?

- A) A mode that limits Claude to read-only operations, preventing file writes
- B) A mode where Claude outlines its intended actions before executing them, allowing human review and approval of the plan before execution
- C) A scheduled execution mode for running Claude tasks at specified times
- D) A mode that reduces token usage by planning ahead

✅ **B** — Plan Mode supports human-in-the-loop review before irreversible actions are taken.

---

**Q34.** Where should custom slash commands be stored in a Claude Code project to make them available to all team members?

- A) In `~/.claude/commands/` (user-level)
- B) In `.claude/commands/` within the project directory (project-level, committed to VCS)
- C) In the root `CLAUDE.md` file as inline command definitions
- D) In the MCP server configuration

✅ **B** — Project-level `.claude/commands/` are shared via version control; user-level is personal.

---

**Q35.** A CLAUDE.md rule states "Always write TypeScript with strict null checks." A subdirectory `.claude/rules/legacy-code.md` states "Null checks are optional in this directory." Which rule applies in that subdirectory?

- A) The root CLAUDE.md rule always takes precedence
- B) The subdirectory rule overrides the root rule within that directory's scope
- C) Both rules apply simultaneously, causing a conflict Claude must resolve
- D) Neither rule applies; conflicting rules are ignored

✅ **B** — More specific (lower-level) CLAUDE.md rules override higher-level rules within their scope.

---

**Q36.** You want to restrict Claude Code to only use the Read, Grep, and Glob tools during a code analysis Skill, preventing any file writes. How do you configure this?

- A) Add "Do not write files" to the Skill's description
- B) Specify `allowed-tools: [Read, Grep, Glob]` in the Skill's frontmatter
- C) Remove the Write and Edit tools from the global Claude Code installation
- D) Use a PreToolUse hook to block Write calls

✅ **B** — `allowed-tools` in Skill frontmatter restricts which tools can be invoked during that Skill.

---

**Q37.** A team is using Claude Code to generate automated pull request comments. The output must be machine-readable by their CI system. What format should Claude be configured to output?

- A) Markdown with structured headers
- B) Plain text with key-value pairs
- C) Validated JSON matching a predefined schema
- D) YAML for compatibility with GitHub Actions

✅ **C** — Validated JSON with a defined schema ensures reliable machine-readable output for CI systems.

---

**Q38.** Which CLAUDE.md entry is the most effective for maintaining consistent commit message style across a team?

- A) "Write good commit messages"
- B) "Follow conventional commit format: `<type>(<scope>): <subject>`, where type is one of: feat, fix, docs, refactor, test, chore"
- C) "Commit messages should describe what changed"
- D) "Use imperative mood in commit messages"

✅ **B** — Explicit format specifications with examples produce consistent, predictable output.

---

**Q39.** A Claude Code skill is producing very long chain-of-thought outputs that are consuming the main session's context. What is the correct fix?

- A) Set a lower `max_tokens` for the Skill
- B) Add `context: fork` to the Skill's frontmatter so it runs in an isolated subagent
- C) Disable the Skill's internal reasoning with a system prompt instruction
- D) Reduce the Skill's description length

✅ **B** — `context: fork` isolates verbose Skills from the main context window.

---

### 🔷 DOMAIN 4 — Prompt Engineering & Structured Output (12 Questions)

**Q40.** You are building a data extraction pipeline that must return a JSON object with a `notes` field that is sometimes absent in the source document. How should the JSON schema define this field to minimize hallucinations?

- A) Mark the field as `required` with a default value of `"N/A"`
- B) Mark the field as nullable (`"type": ["string", "null"]`) and not required, so Claude can return `null` when absent
- C) Omit the field from the schema entirely
- D) Use a `"description"` annotation telling Claude to guess if the value is missing

✅ **B** — Nullable fields allow Claude to return `null` for absent values rather than fabricating content.

---

**Q41.** A prompt for contract analysis is producing inconsistent outputs — some responses include risk ratings, others do not. What is the most effective fix?

- A) Increase the temperature to encourage more complete responses
- B) Add explicit output criteria in the prompt: "For each clause identified, you MUST include: clause_type, risk_level (high/medium/low), and a one-sentence rationale."
- C) Use a larger model to improve instruction following
- D) Add "Be thorough and complete" to the system prompt

✅ **B** — Explicit, enumerated output criteria with required fields eliminate inconsistency more reliably than general instructions.

---

**Q42.** You need Claude to classify customer feedback into categories that may not always be clearly defined by the input. Which prompting technique is MOST effective?

- A) Zero-shot prompting with a brief category list
- B) Few-shot prompting with 2-3 labeled examples for each category, including edge cases that demonstrate classification decisions for ambiguous inputs
- C) Chain-of-thought prompting asking Claude to explain its reasoning
- D) Increasing max_tokens to give Claude more room to reason

✅ **B** — Few-shot examples, especially with edge cases, most effectively handle ambiguous classification.

---

**Q43.** A validation-retry loop has been running for 5 iterations and the output still fails JSON schema validation. What should the loop do?

- A) Continue retrying indefinitely until a valid response is produced
- B) Lower the validation requirements to accept the partial output
- C) Trigger the escalation path: log the failure, alert a human reviewer, and return a structured error to the caller
- D) Switch to a different model for the next retry

✅ **C** — Retry loops must have a defined maximum and an escalation path to avoid infinite loops.

---

**Q44.** Which of the following is the BEST way to use the Messages Batch API?

- A) For real-time customer-facing requests that need sub-second response times
- B) For processing large volumes of independent, non-time-sensitive tasks (e.g., classifying 50,000 documents overnight) at reduced cost
- C) For streaming long-form responses to users
- D) For multi-turn conversations that require maintaining session state

✅ **B** — Batch API is designed for high-volume, asynchronous, non-urgent workloads.

---

**Q45.** A multi-pass review system uses three independent Claude instances to review the same document and flag issues. The final output is the union of issues found by at least two of the three reviewers. What does this architecture improve?

- A) Processing speed by parallelizing review
- B) Output reliability by using majority-consensus to reduce false positives from any single instance
- C) Cost efficiency by reducing the total number of tokens processed
- D) Context management by splitting the document across three instances

✅ **B** — Multi-instance review with consensus voting is a reliability pattern that filters single-instance errors.

---

**Q46.** When designing a structured output schema for an agent that extracts invoice data, which field definition is MOST likely to cause hallucinations?

- A) `"invoice_number": { "type": "string" }`
- B) `"total_amount": { "type": "number" }`
- C) `"payment_terms": { "type": "string", "description": "Infer payment terms if not explicitly stated" }`
- D) `"issue_date": { "type": ["string", "null"] }`

✅ **C** — Instructing Claude to "infer" or fill in missing values encourages hallucination. Use nullable types instead.

---

**Q47.** What is the primary risk of using `tool_use` for structured output versus asking Claude to output raw JSON in a text response?

- A) `tool_use` is slower and more expensive
- B) Raw JSON in text responses is less reliable because Claude may add preamble, commentary, or malformed JSON; `tool_use` enforces schema adherence at the API level
- C) `tool_use` does not support nested JSON objects
- D) Raw JSON is more readable and easier to parse

✅ **B** — `tool_use` with a defined input schema is the production-grade approach for reliable structured output.

---

**Q48.** A prompt asks Claude to "analyze this contract and provide your assessment." Outputs are inconsistent in length and structure. What is the BEST prompt engineering fix?

- A) Ask Claude to "be more detailed and consistent"
- B) Define explicit review criteria: "For each section: (1) identify the key obligation, (2) assign a risk score 1-5, (3) flag any ambiguous language. Return results as a JSON array."
- C) Add "Think step by step" to the prompt
- D) Provide a longer, more detailed contract as context

✅ **B** — Explicit criteria with enumerated steps and a defined output format produce consistent results.

---

**Q49.** You want Claude to extract data from medical reports. Some reports use the term "BP" while others say "blood pressure." Which technique BEST handles this variation?

- A) Create separate prompts for each terminology variant
- B) Use few-shot examples that demonstrate extraction for both terminology variants, showing Claude the normalization expected
- C) Add a glossary to the system prompt and rely on Claude to apply it
- D) Use keyword detection pre-processing to normalize terms before sending to Claude

✅ **B** — Few-shot examples demonstrating normalization of variants are more reliable than glossaries alone.

---

**Q50.** After implementing a validation-retry loop, you notice Claude is retrying even on responses that are valid but slightly inconsistent in formatting. What should you fix?

- A) Remove the retry loop entirely
- B) Tighten the JSON schema to enforce exact formatting requirements, and ensure the validator only rejects structurally invalid responses, not stylistic variations
- C) Add "Format your responses exactly the same each time" to the system prompt
- D) Switch from JSON schema validation to regex validation

✅ **B** — The validator logic should be precise: reject invalid structure, accept valid-but-varied formatting.

---

**Q51.** A data pipeline extracts entities from news articles and returns them in a JSON array. Occasionally Claude returns valid JSON but with the array wrapped in a markdown code block (` ```json `). What is the most robust fix?

- A) Add "Do not use markdown" to the system prompt
- B) Use `tool_use` with a defined input schema to enforce structured output at the API level, eliminating the possibility of markdown wrapping
- C) Post-process responses to strip markdown code fences before parsing
- D) Switch to a lower temperature setting

✅ **B** — `tool_use` prevents markdown wrapping entirely; post-processing is a fragile workaround.

---

### 🔷 DOMAIN 5 — Context Management & Reliability (9 Questions)

**Q52.** A long-running agentic session has accumulated 180,000 tokens of conversation history. Claude is approaching its context limit. What is the BEST strategy to preserve session continuity?

- A) Start a fresh session and re-inject the original task description
- B) Implement progressive summarization: replace older conversation turns with a structured summary block containing key decisions, facts, and current state
- C) Increase the model's temperature to improve creative problem-solving with less context
- D) Truncate the oldest messages until the context fits

✅ **B** — Progressive summarization preserves critical information while reducing token count.

---

**Q53.** An agent frequently loses track of customer account details mentioned early in a long conversation. What architectural pattern BEST addresses this?

- A) Increase the context window size by switching models
- B) Extract and store key facts (customer name, account ID, open issues) in an immutable structured "case facts" block pinned at the top of the context
- C) Periodically ask the user to re-confirm their account details
- D) Enable Claude's implicit long-term memory

✅ **B** — Immutable structured fact blocks at the top of context ensure critical information persists.

---

**Q54.** A subagent is processing a very long document for summarization. After processing, it needs to pass results to a coordinator agent. What should the subagent return?

- A) The full document text with annotations
- B) A structured summary containing only the key facts, decisions, and action items needed by the coordinator
- C) A list of all tokens processed for billing purposes
- D) The document split into chunks for the coordinator to reassemble

✅ **B** — Subagents should return minimal, structured outputs to prevent coordinator context overload.

---

**Q55.** Which of the following scenarios represents a valid escalation trigger in a production agentic system?

- A) The agent has made more than 10 tool calls in a single session
- B) The agent's internal confidence score drops below a threshold
- C) The user explicitly requests to speak with a human representative
- D) The agent detects negative sentiment in the user's message

✅ **C** — Explicit user requests for human assistance are always valid escalation triggers. Heuristics like confidence scores and sentiment are unreliable.

---

**Q56.** How should errors be propagated in a multi-agent system where a subagent fails during execution?

- A) The subagent silently ignores the error and returns an empty result
- B) The subagent returns a structured error response with error type, context, and suggested action, which the coordinator uses to decide whether to retry, reassign, or escalate
- C) The coordinator automatically retries the subagent 3 times regardless of error type
- D) All errors trigger a full session reset

✅ **B** — Structured error propagation enables intelligent coordinator-level decision-making.

---

**Q57.** A production agentic system processes sensitive legal documents. You need to track which source documents contributed to each claim in the final output. What pattern addresses this?

- A) Include a disclaimer that outputs may not be 100% accurate
- B) Implement information provenance tracking: each extracted claim is tagged with its source document ID and relevant passage reference
- C) Use a higher-quality model to reduce errors
- D) Add a human review step after every 10 documents processed

✅ **B** — Provenance tracking is the architectural pattern for source attribution in high-stakes document processing.

---

**Q58.** A scratchpad file strategy is recommended for which type of agentic task?

- A) Short 2-3 turn conversations with simple factual queries
- B) Long-running sessions where the agent needs to accumulate intermediate results, track progress, and reference prior findings without filling the context window
- C) Real-time streaming responses to end users
- D) Single-tool MCP integrations with stateless operations

✅ **B** — Scratchpad files externalize intermediate state, preserving context window capacity for reasoning.

---

**Q59.** An agent is processing a 500-page document. Rather than loading the entire document into context at once, what is the BEST approach?

- A) Ask the user to split the document manually before submission
- B) Use a subagent with `context: fork` to process document sections independently and return structured summaries to the coordinator
- C) Set `max_tokens` to a very high value and load everything at once
- D) Use the Batch API to process the document in a single request

✅ **B** — Subagents with isolated contexts are the correct pattern for large document processing.

---

**Q60.** Which combination of signals should an agent use to determine whether to proceed autonomously or escalate to a human?

- A) The agent's self-assessed confidence score and elapsed processing time
- B) Whether the task falls within the agent's authorized scope, whether the user has explicitly requested human involvement, and whether the agent has reached a genuine capability limit
- C) The number of previous escalations in the session and user sentiment score
- D) The cost of the current API calls relative to session budget

✅ **B** — Authorized scope, explicit user preference, and genuine capability limits are the three reliable escalation signals.

---

## 📊 Domain Distribution Summary

| Domain | Weight | Questions |
|---|---|---|
| D1: Agentic Architecture & Orchestration | 27% | Q1–Q16 (16) |
| D2: Tool Design & MCP Integration | 18% | Q17–Q27 (11) |
| D3: Claude Code Configuration & Workflows | 20% | Q28–Q39 (12) |
| D4: Prompt Engineering & Structured Output | 20% | Q40–Q51 (12) |
| D5: Context Management & Reliability | 15% | Q52–Q60 (9) |

---

**Key exam facts to know:** The passing score is 720 on a scaled 100–1,000 range, the exam is closed-book with no AI assistance allowed, and each question has one correct answer and three distractor responses designed to be genuinely tempting for candidates with partial knowledge. [Claudecertifiedarchitect](https://claudecertifiedarchitect.net/) The exam is built around 6 production scenarios, with 4 randomly selected on exam day — so study all six: Customer Support Agent, Code Generation with Claude Code, Multi-Agent Research System, Developer Productivity, Claude Code for CI/CD, and Structured Data Extraction. [Medium](https://dynamicbalaji.medium.com/claude-certified-architect-foundations-certification-preparation-guide-c70546b51f51)