package usecase

type Step struct {
	Content string
	Headers map[string]string
}

type Usecase struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Steps       []Step `json:"-"`
	StepCount   int    `json:"steps"`
}

var Registry = []Usecase{
	{
		ID:          "full_refund",
		Name:        "Full Refund — Defective Product",
		Description: "Customer specifies a defective product and receives an immediate full refund.",
		Steps: []Step{
			{Content: "I'd like a refund for the wireless headphones I bought last week"},
			{Content: "They're defective — the left earcup makes a buzzing noise"},
		},
	},
	{
		ID:          "refund_no_product",
		Name:        "Refund — Product Not Specified",
		Description: "Customer asks for a refund without naming a product. Agent looks up orders and asks to confirm.",
		Steps: []Step{
			{Content: "Hi, I need to return something for a refund"},
			{Content: "The headphones from order 101 please"},
			{Content: "They're defective"},
		},
	},
	{
		ID:          "complaint_then_refund",
		Name:        "Complaint Turns Into Refund",
		Description: "Customer complains about product quality. After clarification, a refund is processed.",
		Steps: []Step{
			{Content: "The t-shirts I ordered are nothing like the pictures on your website"},
			{Content: "Yes, I'd like a refund please"},
			{Content: "They don't match the description at all"},
		},
	},
	{
		ID:          "feedback_only",
		Name:        "Feedback Only — No Refund",
		Description: "Customer provides negative feedback but explicitly declines a refund.",
		Steps: []Step{
			{Content: "I'm not happy with the meal kit I received"},
			{Content: "No, I don't need a refund. Just wanted to let you know the quality was below expectations"},
			{Content: "That's all, thanks"},
		},
	},
	{
		ID:          "refund_denied",
		Name:        "Refund Denied by Policy",
		Description: "Customer wants to return clothing due to change of mind. Policy denies the refund.",
		Steps: []Step{
			{Content: "I want to return the t-shirt I bought"},
			{Content: "I just changed my mind about the color"},
			{Content: "Okay, I understand"},
		},
	},
	{
		ID:          "partial_refund_accepted",
		Name:        "Partial Refund — Accepted",
		Description: "Customer returns electronics (change of mind). Offered a 70% partial refund and accepts.",
		Steps: []Step{
			{Content: "I want to return my keyboard"},
			{Content: "Changed my mind"},
			{Content: "Yes, that's fine, I'll take the partial refund"},
		},
	},
	{
		ID:          "partial_refund_declined",
		Name:        "Partial Refund — Declined → Escalation",
		Description: "Customer refuses the partial refund offer. Agent escalates to human support.",
		Steps: []Step{
			{Content: "I want to return my keyboard"},
			{Content: "I just changed my mind, nothing wrong with it"},
			{Content: "No, I think I deserve a full refund"},
		},
	},
	{
		ID:          "escalate_by_policy",
		Name:        "Escalation — Policy Requires Human Review",
		Description: "Software refunds always require human review. Agent escalates immediately.",
		Steps: []Step{
			{Content: "I need a refund for the photo editing software"},
			{Content: "It crashes every time I try to export"},
		},
	},
	{
		ID:          "escalate_by_error",
		Name:        "Escalation — System Error",
		Description: "A simulated backend error forces the agent to escalate to human support.",
		Steps: []Step{
			{Content: "I want a refund for my running shoes", Headers: map[string]string{"X-Demo-Force-Error": "true"}},
			{Content: "They fell apart after one run", Headers: map[string]string{"X-Demo-Force-Error": "true"}},
		},
	},
	{
		ID:          "subscription_trial_refund",
		Name:        "Subscription — Trial Window Refund",
		Description: "Customer cancels a subscription within the 7-day trial and gets a full refund.",
		Steps: []Step{
			{Content: "I want to cancel CloudSync Pro, I signed up 2 days ago"},
			{Content: "I just changed my mind"},
		},
	},
	{
		ID:          "subscription_late_cancel",
		Name:        "Subscription — Outside Window (Denied)",
		Description: "Customer tries to cancel a subscription months later. Policy denies the refund.",
		Steps: []Step{
			{Content: "Cancel my CloudSync Pro subscription and refund me"},
			{Content: "I've had it for a few months but never really used it"},
			{Content: "Fine, I understand"},
		},
	},
}

func init() {
	for i := range Registry {
		Registry[i].StepCount = len(Registry[i].Steps)
	}
}

func FindByID(id string) *Usecase {
	for i := range Registry {
		if Registry[i].ID == id {
			return &Registry[i]
		}
	}
	return nil
}
