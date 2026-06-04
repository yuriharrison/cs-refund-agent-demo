package testutil

// ScenarioSnapshots returns mock OpenAI request/response pairs for each usecase.
// Snapshots drive tool-call sequences; regenerate with `make test-refresh` for live API recordings.
func ScenarioSnapshots() map[string][]RequestResponsePair {
	return map[string][]RequestResponsePair{
		"scenario_refund_product_specified":  refundProductSpecified(),
		"scenario_refund_no_product":         refundNoProduct(),
		"scenario_feedback_only":             feedbackOnly(),
		"scenario_complaint_then_refund":     complaintThenRefund(),
		"scenario_refund_denied":             refundDenied(),
		"scenario_full_refund_auto":          fullRefundAuto(),
		"scenario_partial_refund_declined":   partialRefundDeclined(),
		"scenario_partial_refund_accepted":   partialRefundAccepted(),
		"scenario_escalate_by_policy":        escalateByPolicy(),
		"scenario_escalate_by_error":         escalateByError(),
		"scenario_subscription_trial_refund": subscriptionTrialRefund(),
		"scenario_subscription_late_cancel":  subscriptionLateCancel(),
	}
}

func refundProductSpecified() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I'd be happy to help with your refund request. Could you tell me what issue you're experiencing with the wireless headphones?"),
		pairToolCall("call_rps_policy", "get_refund_policy", `{"product_type":"electronics","condition":"defective"}`),
		pairToolCall("call_rps_refund", "issue_refund", `{"order_item_id":1,"refund_type":"full","amount":149.99,"reason":"defective — buzzing noise in left earcup"}`),
		pairContent("I've processed a full refund of $149.99 for your wireless headphones. You should see it on your original payment method within 5-7 business days."),
	}
}

func refundNoProduct() []RequestResponsePair {
	return []RequestResponsePair{
		pairToolCall("call_rnp_lookup", "lookup_customer_orders", `{"limit":5}`),
		pairContent("I can see your recent orders. Which item would you like to return for a refund?"),
		pairContent("Thank you. Could you tell me what's wrong with the headphones from order 101?"),
		pairToolCall("call_rnp_policy", "get_refund_policy", `{"product_type":"electronics","condition":"defective"}`),
		pairToolCall("call_rnp_refund", "issue_refund", `{"order_item_id":1,"refund_type":"full","amount":149.99,"reason":"defective"}`),
		pairContent("I've issued a full refund of $149.99 for the headphones from order 101."),
	}
}

func feedbackOnly() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I'm sorry to hear the meal kit didn't meet your expectations. Would you like me to help you with a refund?"),
		pairContent("Thank you for sharing your feedback. I've noted your concerns about the meal kit quality and will pass this along to our team."),
		pairContent("You're welcome! If you need anything else in the future, don't hesitate to reach out. Have a great day!"),
	}
}

func complaintThenRefund() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I'm sorry the t-shirts didn't match what you expected. Would you like to proceed with a refund?"),
		pairContent("I'd be happy to help. Could you describe how the t-shirts differ from what was shown on the website?"),
		pairToolCall("call_ctr_policy", "get_refund_policy", `{"product_type":"clothing","condition":"not_as_described"}`),
		pairToolCall("call_ctr_refund", "issue_refund", `{"order_item_id":2,"refund_type":"full","amount":69.98,"reason":"not as described — t-shirts don't match website photos"}`),
		pairContent("I've processed a full refund of $69.98 for the t-shirts. You should see it within 5-7 business days."),
	}
}

func refundDenied() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I can help with that. Could you tell me the reason you'd like to return the t-shirt?"),
		pairToolCall("call_rd_policy", "get_refund_policy", `{"product_type":"clothing","condition":"change_of_mind"}`),
		pairContent("Unfortunately, our policy doesn't allow refunds for change-of-mind returns on clothing items due to hygiene regulations. I'm sorry I can't process a refund in this case."),
		pairContent("I understand. Thank you for your understanding. Please let me know if there's anything else I can help with."),
	}
}

func fullRefundAuto() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I'm sorry to hear about your headphones. Could you tell me more about what's wrong with them?"),
		pairToolCall("call_fra_policy", "get_refund_policy", `{"product_type":"electronics","condition":"defective"}`),
		pairToolCall("call_fra_refund", "issue_refund", `{"order_item_id":1,"refund_type":"full","amount":149.99,"reason":"defective — stopped working after 2 days"}`),
		pairContent("I've processed a full refund of $149.99 for your headphones. You should see it on your original payment method within 5-7 business days."),
	}
}

func partialRefundDeclined() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I can help with your keyboard return. What's the reason you'd like to return it?"),
		pairToolCall("call_prd_policy", "get_refund_policy", `{"product_type":"electronics","condition":"change_of_mind"}`),
		pairContent("Based on our policy, I can offer a partial refund of 70% ($55.99) for change-of-mind returns on electronics. Would you like to accept this offer?"),
		pairToolCall("call_prd_escalate", "escalate_to_human", `{"reason":"customer declined partial refund offer for keyboard return"}`),
		pairContent("I understand you'd prefer a full refund. I've escalated your case to a human support agent who will review it shortly."),
	}
}

func partialRefundAccepted() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I can help with your keyboard return. What's the reason you'd like to return it?"),
		pairToolCall("call_pra_policy", "get_refund_policy", `{"product_type":"electronics","condition":"change_of_mind"}`),
		pairContent("Based on our policy, I can offer a partial refund of 70% ($55.99) for change-of-mind returns. Would you like to accept?"),
		pairToolCall("call_pra_refund", "issue_refund", `{"order_item_id":6,"refund_type":"partial","amount":55.99,"reason":"change of mind — customer accepted partial refund"}`),
		pairContent("I've processed a partial refund of $55.99 for your keyboard. You should see it within 5-7 business days."),
	}
}

func escalateByPolicy() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I can help with your software refund request. Could you describe the issue you're experiencing?"),
		pairToolCall("call_ebp_policy", "get_refund_policy", `{"product_type":"software","condition":"defective"}`),
		pairToolCall("call_ebp_escalate", "escalate_to_human", `{"reason":"software refund requires human review per policy"}`),
		pairContent("Software refunds require review by our support team. I've connected you with a specialist who will assist you shortly."),
	}
}

func escalateByError() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I'm sorry to hear about your running shoes. Could you tell me more about what happened?"),
		pairToolCall("call_ebe_policy", "get_refund_policy", `{"product_type":"clothing","condition":"defective"}`),
		pairToolCall("call_ebe_escalate", "escalate_to_human", `{"reason":"unable to determine refund policy due to system error"}`),
		pairContent("I encountered an issue checking the refund policy. I've escalated your case to a human agent who will help you right away."),
	}
}

func subscriptionTrialRefund() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I can help you cancel CloudSync Pro. Could you confirm the reason for cancellation?"),
		pairToolCall("call_str_policy", "get_refund_policy", `{"product_type":"subscription","condition":"change_of_mind"}`),
		pairToolCall("call_str_refund", "issue_refund", `{"order_item_id":5,"refund_type":"full","amount":119.99,"reason":"change of mind within trial window"}`),
		pairContent("I've processed a full refund of $119.99 for your CloudSync Pro subscription and cancelled your plan."),
	}
}

func subscriptionLateCancel() []RequestResponsePair {
	return []RequestResponsePair{
		pairContent("I can help with your CloudSync Pro cancellation request. Could you tell me more about why you'd like to cancel?"),
		pairToolCall("call_slc_policy", "get_refund_policy", `{"product_type":"subscription","condition":"any"}`),
		pairContent("I'm sorry, but our policy doesn't allow refunds for subscriptions outside the trial or renewal windows. Your subscription will remain active until the end of the billing period."),
		pairContent("I understand. If you change your mind or need anything else, please don't hesitate to reach out."),
	}
}

func WriteAllScenarioSnapshots() error {
	for name, pairs := range ScenarioSnapshots() {
		if err := WriteSnapshotFile(name, pairs); err != nil {
			return err
		}
	}
	return nil
}
