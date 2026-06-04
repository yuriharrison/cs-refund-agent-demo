import { test, expect } from "@playwright/test";
import {
  mockChatMessage,
  mockChatReset,
  mockSSEWithEvents,
  type MockSSEEvent,
} from "./helpers";

const SESSION_ID = "test-session-001";

test.describe("Chat UI — Layout & Structure", () => {
  test.beforeEach(async ({ page }) => {
    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, []);
    await page.goto("/");
  });

  test("renders the chat container with header, message area, and input", async ({
    page,
  }) => {
    await expect(page.getByTestId("chat-container")).toBeVisible();
    await expect(page.getByTestId("chat-header")).toBeVisible();
    await expect(page.getByTestId("message-list")).toBeVisible();
    await expect(page.getByTestId("chat-footer")).toBeVisible();
  });

  test("header shows ShopEase Support title", async ({ page }) => {
    await expect(page.getByText("ShopEase Support").first()).toBeVisible();
  });

  test("shows empty state message when no messages", async ({ page }) => {
    await expect(page.getByTestId("empty-state")).toBeVisible();
    await expect(
      page.getByText("I'm your AI support agent"),
    ).toBeVisible();
  });

  test("shows chat input with placeholder", async ({ page }) => {
    const input = page.getByTestId("chat-input");
    await expect(input).toBeVisible();
    await expect(input).toHaveAttribute("placeholder", "Type your message...");
  });

  test("send button is disabled when input is empty", async ({ page }) => {
    await expect(page.getByTestId("send-button")).toBeDisabled();
  });

  test("send button is enabled when input has text", async ({ page }) => {
    await page.getByTestId("chat-input").fill("Hello");
    await expect(page.getByTestId("send-button")).toBeEnabled();
  });

  test("token counter shows zero initial state", async ({ page }) => {
    const counter = page.getByTestId("token-counter");
    await expect(counter).toBeVisible();
    await expect(counter).toContainText("Tokens: 0");
  });

  test("reset button is initially disabled", async ({ page }) => {
    await expect(page.getByTestId("reset-button")).toBeDisabled();
  });
});

test.describe("Chat UI — Sending Messages", () => {
  test("sends a message and shows customer bubble", async ({ page }) => {
    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, []);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("I need help with a refund");
    await page.getByTestId("send-button").click();

    const customerMsg = page.getByTestId("message-customer");
    await expect(customerMsg).toBeVisible();
    await expect(customerMsg).toContainText("I need help with a refund");
  });

  test("empty state disappears after sending a message", async ({ page }) => {
    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, []);
    await page.goto("/");

    await expect(page.getByTestId("empty-state")).toBeVisible();

    await page.getByTestId("chat-input").fill("Hello");
    await page.getByTestId("send-button").click();

    await expect(page.getByTestId("empty-state")).not.toBeVisible();
  });

  test("input clears after sending", async ({ page }) => {
    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, []);
    await page.goto("/");

    const input = page.getByTestId("chat-input");
    await input.fill("Hello");
    await page.getByTestId("send-button").click();

    await expect(input).toHaveValue("");
  });

  test("input is disabled while processing", async ({ page }) => {
    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, []);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("Hello");
    await page.getByTestId("send-button").click();

    await expect(page.getByTestId("chat-input")).toBeDisabled();
  });

  test("can send with Enter key", async ({ page }) => {
    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, []);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("Hello there");
    await page.getByTestId("chat-input").press("Enter");

    await expect(page.getByTestId("message-customer")).toBeVisible();
    await expect(page.getByTestId("message-customer")).toContainText(
      "Hello there",
    );
  });
});

test.describe("Chat UI — SSE Event Rendering", () => {
  const agentEvents: MockSSEEvent[] = [
    { type: "agent_thinking", data: { content: "Processing your request..." } },
    {
      type: "agent_message",
      data: { content: "I'd be happy to help you with your refund!" },
    },
    {
      type: "token_update",
      data: { prompt_tokens: 500, completion_tokens: 50, total: 550 },
    },
  ];

  test("displays agent message from SSE", async ({ page }) => {
    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, agentEvents);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("Help me");
    await page.getByTestId("send-button").click();

    const agentMsg = page.getByTestId("message-agent");
    await expect(agentMsg).toBeVisible({ timeout: 5000 });
    await expect(agentMsg).toContainText(
      "I'd be happy to help you with your refund!",
    );
  });

  test("updates token counter from SSE events", async ({ page }) => {
    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, agentEvents);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("Help me");
    await page.getByTestId("send-button").click();

    await expect(page.getByTestId("token-counter")).toContainText("550", {
      timeout: 5000,
    });
  });

  test("displays tool call debug cards from SSE", async ({ page }) => {
    const toolEvents: MockSSEEvent[] = [
      { type: "agent_thinking", data: { content: "Processing..." } },
      {
        type: "tool_call_start",
        data: { tool: "lookup_customer_orders", arguments: { limit: 5 } },
      },
      {
        type: "tool_call_result",
        data: {
          tool: "lookup_customer_orders",
          result: [{ order_id: 101 }],
          duration_ms: 12,
        },
      },
      { type: "agent_message", data: { content: "Here are your orders." } },
    ];

    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, toolEvents);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("Show my orders");
    await page.getByTestId("send-button").click();

    const debugStart = page.getByTestId("debug-start").first();
    await expect(debugStart).toBeVisible({ timeout: 5000 });
    await expect(debugStart).toContainText("Looking up orders");

    const debugResult = page.getByTestId("debug-result").first();
    await expect(debugResult).toBeVisible();
    await expect(debugResult).toContainText("12ms");
  });

  test("debug cards are collapsed by default and expand on click", async ({
    page,
  }) => {
    const toolEvents: MockSSEEvent[] = [
      {
        type: "tool_call_start",
        data: { tool: "get_refund_policy", arguments: { product_type: "electronics" } },
      },
      { type: "agent_message", data: { content: "Done" } },
    ];

    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, toolEvents);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("Check policy");
    await page.getByTestId("send-button").click();

    const debugCard = page.getByTestId("debug-start");
    await expect(debugCard).toBeVisible({ timeout: 5000 });

    // Content should not be visible (collapsed)
    await expect(debugCard.locator("pre")).not.toBeVisible();

    // Click to expand
    await debugCard.getByTestId("debug-toggle").click();

    // Content should now be visible
    await expect(debugCard.locator("pre")).toBeVisible();
    await expect(debugCard.locator("pre")).toContainText("electronics");
  });

  test("displays system refund confirmation card", async ({ page }) => {
    const refundEvents: MockSSEEvent[] = [
      { type: "agent_thinking", data: { content: "Processing..." } },
      { type: "agent_message", data: { content: "Refund processed!" } },
      {
        type: "system_confirmation",
        data: {
          action: "refund_issued",
          details: {
            refund_id: "REF-001",
            amount: 149.99,
            type: "full",
            order_item_id: "OI-101",
          },
        },
      },
    ];

    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, refundEvents);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("Refund my headphones");
    await page.getByTestId("send-button").click();

    const refundCard = page.getByTestId("system-refund");
    await expect(refundCard).toBeVisible({ timeout: 5000 });
    await expect(refundCard).toContainText("Refund Confirmed");
    await expect(refundCard).toContainText("$149.99");
    await expect(refundCard).toContainText("full");
  });

  test("displays system escalation card", async ({ page }) => {
    const escalationEvents: MockSSEEvent[] = [
      { type: "agent_thinking", data: { content: "Processing..." } },
      {
        type: "agent_message",
        data: { content: "I'm connecting you with a specialist." },
      },
      {
        type: "system_escalation",
        data: {
          reason: "Software refund requires human review",
          human_agent_name: "Alex",
        },
      },
    ];

    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, escalationEvents);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("Refund my software");
    await page.getByTestId("send-button").click();

    const escalationCard = page.getByTestId("system-escalation");
    await expect(escalationCard).toBeVisible({ timeout: 5000 });
    await expect(escalationCard).toContainText("Escalated to Human Agent");
    await expect(escalationCard).toContainText("Alex");
  });

  test("displays human agent message", async ({ page }) => {
    const humanEvents: MockSSEEvent[] = [
      {
        type: "system_escalation",
        data: { reason: "needs review", human_agent_name: "Alex" },
      },
      {
        type: "human_message",
        data: {
          content:
            "Hi, I'm Alex from the support team. I've reviewed your case.",
          agent_name: "Alex",
        },
      },
    ];

    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, humanEvents);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("I need help");
    await page.getByTestId("send-button").click();

    const humanMsg = page.getByTestId("message-human");
    await expect(humanMsg).toBeVisible({ timeout: 5000 });
    await expect(humanMsg).toContainText("Alex");
    await expect(humanMsg).toContainText("I've reviewed your case");
    await expect(humanMsg).toContainText("Support Agent");
  });

  test("displays error card on error event", async ({ page }) => {
    const errorEvents: MockSSEEvent[] = [
      { type: "error", data: { message: "Something went wrong" } },
    ];

    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, errorEvents);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("Trigger error");
    await page.getByTestId("send-button").click();

    const errorCard = page.getByTestId("system-error");
    await expect(errorCard).toBeVisible({ timeout: 5000 });
    await expect(errorCard).toContainText("Error");
    await expect(errorCard).toContainText("Something went wrong");
  });
});

test.describe("Chat UI — Session Management", () => {
  test("session ID appears in header after first message", async ({
    page,
  }) => {
    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, [
      { type: "agent_message", data: { content: "Hi!" } },
    ]);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("Hello");
    await page.getByTestId("send-button").click();

    await expect(page.getByTestId("session-id")).toContainText(
      SESSION_ID.slice(0, 8),
      { timeout: 5000 },
    );
  });

  test("reset clears all messages and starts fresh", async ({ page }) => {
    const newSessionId = "new-session-002";
    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, [
      { type: "agent_message", data: { content: "Hi there!" } },
    ]);
    await page.goto("/");

    // Send a message
    await page.getByTestId("chat-input").fill("Hello");
    await page.getByTestId("send-button").click();

    await expect(page.getByTestId("message-customer")).toBeVisible();

    // Set up reset mock
    await mockChatReset(page, newSessionId);
    // Re-mock SSE for new session
    await mockSSEWithEvents(page, []);

    // Reset
    await page.getByTestId("reset-button").click();

    // Messages should be cleared
    await expect(page.getByTestId("empty-state")).toBeVisible({ timeout: 5000 });
    await expect(page.getByTestId("message-customer")).not.toBeVisible();
  });
});

test.describe("Chat UI — Full Refund Flow (mocked)", () => {
  test("complete refund flow: customer → agent → tool calls → refund confirmation", async ({
    page,
  }) => {
    const fullFlowEvents: MockSSEEvent[] = [
      { type: "agent_thinking", data: { content: "Processing your request..." } },
      {
        type: "tool_call_start",
        data: {
          tool: "lookup_customer_orders",
          arguments: { limit: 5 },
        },
      },
      {
        type: "tool_call_result",
        data: {
          tool: "lookup_customer_orders",
          result: {
            orders: [{ id: 101, items: [{ name: "Wireless Headphones" }] }],
          },
          duration_ms: 8,
        },
      },
      {
        type: "tool_call_start",
        data: {
          tool: "get_refund_policy",
          arguments: { product_type: "electronics", condition: "defective" },
        },
      },
      {
        type: "tool_call_result",
        data: {
          tool: "get_refund_policy",
          result: { action: "full_refund", window_days: 30 },
          duration_ms: 3,
        },
      },
      {
        type: "tool_call_start",
        data: {
          tool: "issue_refund",
          arguments: {
            order_item_id: 1,
            refund_type: "full",
            amount: 149.99,
            reason: "defective",
          },
        },
      },
      {
        type: "tool_call_result",
        data: {
          tool: "issue_refund",
          result: { refund_id: 1, status: "approved", amount: 149.99 },
          duration_ms: 5,
        },
      },
      {
        type: "agent_message",
        data: {
          content:
            "Great news! I've processed a full refund of $149.99 for your Wireless Headphones.",
        },
      },
      {
        type: "system_confirmation",
        data: {
          action: "refund_issued",
          details: {
            refund_id: "1",
            amount: 149.99,
            type: "full",
            order_item_id: "1",
          },
        },
      },
      {
        type: "token_update",
        data: { prompt_tokens: 1200, completion_tokens: 85, total: 1285 },
      },
    ];

    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, fullFlowEvents);
    await page.goto("/");

    // Customer sends message
    await page.getByTestId("chat-input").fill(
      "My headphones are broken, I need a refund",
    );
    await page.getByTestId("send-button").click();

    // Customer message visible
    await expect(page.getByTestId("message-customer")).toContainText(
      "headphones are broken",
    );

    // Agent response appears
    const agentMsg = page.getByTestId("message-agent");
    await expect(agentMsg).toBeVisible({ timeout: 5000 });
    await expect(agentMsg).toContainText("$149.99");

    // Tool call debug cards appear
    const debugCards = page.getByTestId("debug-start");
    await expect(debugCards).toHaveCount(3);

    // Refund confirmation card appears
    const refundCard = page.getByTestId("system-refund");
    await expect(refundCard).toBeVisible();
    await expect(refundCard).toContainText("$149.99");
    await expect(refundCard).toContainText("Refund Confirmed");

    // Token counter updated
    await expect(page.getByTestId("token-counter")).toContainText("1,285");
  });
});

test.describe("Chat UI — Escalation Flow (mocked)", () => {
  test("escalation flow: agent → escalation card → human agent message", async ({
    page,
  }) => {
    const escalationEvents: MockSSEEvent[] = [
      { type: "agent_thinking", data: { content: "Processing..." } },
      {
        type: "tool_call_start",
        data: {
          tool: "get_refund_policy",
          arguments: { product_type: "software", condition: "defective" },
        },
      },
      {
        type: "tool_call_result",
        data: {
          tool: "get_refund_policy",
          result: { action: "escalate", notes: "All software refunds require human review" },
          duration_ms: 2,
        },
      },
      {
        type: "tool_call_start",
        data: {
          tool: "escalate_to_human",
          arguments: { reason: "Software refund requires human review" },
        },
      },
      {
        type: "tool_call_result",
        data: {
          tool: "escalate_to_human",
          result: { escalated: true, reason: "Software refund requires human review" },
          duration_ms: 1,
        },
      },
      {
        type: "agent_message",
        data: {
          content:
            "I'm connecting you with a specialist who can help further.",
        },
      },
      {
        type: "system_escalation",
        data: {
          reason: "Software refund requires human review",
          human_agent_name: "Alex",
        },
      },
      {
        type: "human_message",
        data: {
          content:
            "Hi, I'm Alex from the support team. I've reviewed your case and I'm here to help.",
          agent_name: "Alex",
        },
      },
    ];

    await mockChatMessage(page, SESSION_ID);
    await mockSSEWithEvents(page, escalationEvents);
    await page.goto("/");

    await page.getByTestId("chat-input").fill("I need a refund for the photo editing software");
    await page.getByTestId("send-button").click();

    // Agent message visible
    await expect(page.getByTestId("message-agent")).toContainText(
      "connecting you with a specialist",
      { timeout: 5000 },
    );

    // Escalation card visible
    const escalationCard = page.getByTestId("system-escalation");
    await expect(escalationCard).toBeVisible();
    await expect(escalationCard).toContainText("Alex");

    // Human message visible
    const humanMsg = page.getByTestId("message-human");
    await expect(humanMsg).toBeVisible();
    await expect(humanMsg).toContainText("I'm Alex from the support team");
    await expect(humanMsg).toContainText("Support Agent");
  });
});
