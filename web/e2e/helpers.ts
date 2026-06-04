import type { Page, Route } from "@playwright/test";

export interface MockSSEEvent {
  type: string;
  data: Record<string, unknown>;
}

/**
 * Intercepts POST /api/chat/message and responds with a mock session,
 * then feeds SSE events through the already-mocked SSE stream.
 */
export async function mockChatMessage(
  page: Page,
  sessionId: string,
  opts?: { messageId?: string },
) {
  await page.route("**/api/chat/message", async (route) => {
    await route.fulfill({
      status: 202,
      contentType: "application/json",
      body: JSON.stringify({
        session_id: sessionId,
        message_id: opts?.messageId ?? crypto.randomUUID(),
      }),
    });
  });
}

/**
 * Intercepts POST /api/chat/reset and responds with a new session ID.
 */
export async function mockChatReset(
  page: Page,
  newSessionId: string,
) {
  await page.route("**/api/chat/reset", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ session_id: newSessionId }),
    });
  });
}

/**
 * Intercepts GET /api/chat/history and responds with an empty history.
 */
export async function mockChatHistory(page: Page) {
  await page.route("**/api/chat/history**", async (route) => {
    const url = new URL(route.request().url());
    const sessionId = url.searchParams.get("session_id") ?? "mock-session";
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ session_id: sessionId, messages: [] }),
    });
  });
}

/**
 * Intercepts GET /api/chat/stream (SSE) and provides a controller
 * to push events into the stream.
 */
export function mockSSEStream(page: Page): {
  setup: () => Promise<void>;
  pushEvent: (event: MockSSEEvent) => void;
  close: () => void;
} {
  let routeHandler: Route | null = null;
  let encoder: TextEncoder | null = null;
  let chunks: string[] = [];
  let resolveRoute: (() => void) | null = null;

  const setup = async () => {
    await page.route("**/api/chat/stream**", async (route) => {
      routeHandler = route;
      encoder = new TextEncoder();

      for (const chunk of chunks) {
        // Flush any pre-queued events
      }

      await route.fulfill({
        status: 200,
        contentType: "text/event-stream",
        headers: {
          "Cache-Control": "no-cache",
          Connection: "keep-alive",
        },
        body: chunks.join(""),
      });

      resolveRoute?.();
    });
  };

  const pushEvent = (event: MockSSEEvent) => {
    const chunk = `event: ${event.type}\ndata: ${JSON.stringify(event.data)}\n\n`;
    chunks.push(chunk);
  };

  const close = () => {
    routeHandler = null;
  };

  return { setup, pushEvent, close };
}

/**
 * Pre-loads SSE events before the page connects, then fulfills the
 * SSE route with all events at once. This is simpler and more reliable
 * for tests where we know the full event sequence upfront.
 */
export async function mockSSEWithEvents(
  page: Page,
  events: MockSSEEvent[],
) {
  const body = events
    .map((e) => `event: ${e.type}\ndata: ${JSON.stringify(e.data)}\n\n`)
    .join("");

  await page.route("**/api/chat/stream**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "text/event-stream",
      headers: {
        "Cache-Control": "no-cache",
        Connection: "keep-alive",
      },
      body,
    });
  });
}

export async function mockHealthEndpoint(page: Page) {
  await page.route("**/api/health", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ status: "ok", model: "deepseek/deepseek-v4-flash" }),
    });
  });
}
