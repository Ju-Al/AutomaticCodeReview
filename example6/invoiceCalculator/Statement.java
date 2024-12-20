package example6.invoiceCalculator;

import java.text.NumberFormat;
import java.util.Locale;
import java.util.Map;

public class Statement {

    public static String generateStatement(Invoice invoice, Map<String, Play> plays) {
        double totalAmount = 0;
        int volumeCredits = 0;
        StringBuilder result = new StringBuilder("Statement for " + invoice.getCustomer() + "\n");
        NumberFormat currencyFormat = NumberFormat.getCurrencyInstance(Locale.US);

        for (Performance perf : invoice.getPerformances()) {
            Play play = plays.get(perf.getPlayID());
            double thisAmount = 0;

            if (play == null) {
                throw new IllegalArgumentException("Unknown play: " + perf.getPlayID());
            }

            switch (play.getType()) {
                case "tragedy":
                    thisAmount = 40000;
                    if (perf.getAudience() > 30) {
                        thisAmount += 1000 * (perf.getAudience() - 30);
                    }
                    break;

                case "comedy":
                    thisAmount = 30000;
                    if (perf.getAudience() > 20) {
                        thisAmount += 10000 + 500 * (perf.getAudience() - 20);
                    }
                    thisAmount += 300 * perf.getAudience();
                    break;

                default:
                    throw new IllegalArgumentException("Unknown type: " + play.getType());
            }

            // Add volume credits
            volumeCredits += Math.max(perf.getAudience() - 30, 0);

            // Extra credit for every ten comedy attendees
            if ("comedy".equals(play.getType())) {
                volumeCredits += Math.floor(perf.getAudience() / 5);
            }

            // Append performance details to result
            result.append(String.format(" %s: %s (%d seats)\n",
                    play.getName(), currencyFormat.format(thisAmount / 100), perf.getAudience()));
            totalAmount += thisAmount;
        }

        result.append(String.format("Amount owed is %s\n", currencyFormat.format(totalAmount / 100)));
        result.append(String.format("You earned %d credits\n", volumeCredits));
        return result.toString();
    }
}
