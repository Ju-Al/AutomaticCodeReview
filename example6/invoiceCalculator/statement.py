def statement(invoice, plays):
    total_amount = 0
    volume_credits = 0
    result = f"Statement for {invoice['customer']}\n"

    def format_currency(amount):
        return f"${amount:,.2f}"

    for perf in invoice['performances']:
        play = plays[perf['playID']]
        if play is None:
            raise ValueError(f"Unknown play: {perf['playID']}")

        this_amount = 0
        if play['type'] == "tragedy":
            this_amount = 40000
            if perf['audience'] > 30:
                this_amount += 1000 * (perf['audience'] - 30)
        elif play['type'] == "comedy":
            this_amount = 30000
            if perf['audience'] > 20:
                this_amount += 10000 + 500 * (perf['audience'] - 20)
            this_amount += 300 * perf['audience']
        else:
            raise ValueError(f"Unknown type: {play['type']}")

        # Add volume credits
        volume_credits += max(perf['audience'] - 30, 0)

        # Add extra credit for every ten comedy attendees
        if play['type'] == "comedy":
            volume_credits += perf['audience'] // 5

        # Append performance details to result
        result += f" {play['name']}: {format_currency(this_amount / 100)} ({perf['audience']} seats)\n"
        total_amount += this_amount

    result += f"Amount owed is {format_currency(total_amount / 100)}\n"
    result += f"You earned {volume_credits} credits\n"
    return result

plays = {
    "hamlet": {"name": "Hamlet", "type": "tragedy"},
    "as like": {"name": "As You Like It", "type": "comedy"},
    "othello": {"name": "Othello", "type": "tragedy"}
}

invoice = {
    "customer": "BigCo",
    "performances": [
        {"playID": "hamlet", "audience": 55},
        {"playID": "as like", "audience": 35},
        {"playID": "othello", "audience": 40}
    ]
}

print(statement(invoice, plays))