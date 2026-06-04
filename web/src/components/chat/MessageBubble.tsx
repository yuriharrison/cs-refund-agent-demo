import ReactMarkdown from "react-markdown";

interface MessageBubbleProps {
  role: "customer" | "agent";
  content: string;
  streaming?: boolean;
}

export function MessageBubble({ role, content, streaming }: MessageBubbleProps) {
  const isCustomer = role === "customer";

  return (
    <div
      className={`flex animate-fade-in-up ${isCustomer ? "justify-end" : "justify-start"}`}
      data-testid={`message-${role}`}
    >
      <div
        className={`max-w-[80%] rounded-2xl px-4 py-2.5 text-sm leading-relaxed ${
          isCustomer
            ? "bg-indigo-500 text-white"
            : "glass text-slate-100"
        }`}
      >
        {isCustomer ? (
          <p className="whitespace-pre-wrap break-words">{content}</p>
        ) : (
          <div className="prose-chat break-words">
            <ReactMarkdown>{content}</ReactMarkdown>
            {streaming && (
              <span className="ml-0.5 inline-block h-4 w-0.5 animate-pulse bg-slate-300" />
            )}
          </div>
        )}
      </div>
    </div>
  );
}
