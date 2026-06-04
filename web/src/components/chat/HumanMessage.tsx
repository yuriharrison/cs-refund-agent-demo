import ReactMarkdown from "react-markdown";

interface HumanMessageProps {
  content: string;
  agentName: string;
}

export function HumanMessage({ content, agentName }: HumanMessageProps) {
  return (
    <div className="flex animate-fade-in-up justify-start" data-testid="message-human">
      <div className="max-w-[80%]">
        <div className="mb-1 flex items-center gap-2">
          <span className="rounded-full bg-amber-500/20 px-2 py-0.5 text-xs font-medium text-amber-300">
            Support Agent
          </span>
          <span className="text-xs text-amber-400/70">{agentName}</span>
        </div>
        <div className="rounded-2xl border border-amber-500/20 bg-amber-500/10 px-4 py-2.5 text-sm leading-relaxed text-amber-50">
          <div className="prose-chat break-words">
            <ReactMarkdown>{content}</ReactMarkdown>
          </div>
        </div>
      </div>
    </div>
  );
}
