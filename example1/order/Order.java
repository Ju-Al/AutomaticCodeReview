package example1.order;

import java.util.ArrayList;
import java.util.List;

public class Order {
    private final List<String> items;
    private final List<Integer> quantities;
    private final List<Double> prices;
    private String status;

    public Order() {
        this.items = new ArrayList<>();
        this.quantities = new ArrayList<>();
        this.prices = new ArrayList<>();
        this.status = "open";
    }

    public void addItem(String name, int quantity, double price) {
        this.items.add(name);
        this.quantities.add(quantity);
        this.prices.add(price);
    }

    public double totalPrice() {
        double total = 0;
        for (int i = 0; i < this.quantities.size(); i++) {
            total += this.quantities.get(i) * this.prices.get(i);
        }
        return total;
    }

    public void pay(String paymentType, String securityCode) {
        switch (paymentType.toLowerCase()) {
            case "debit":
                System.out.println("Processing debit payment type");
                System.out.println("Verifying security code: " + securityCode);
                this.status = "paid";
                break;

            case "credit":
                System.out.println("Processing credit payment type");
                System.out.println("Verifying security code: " + securityCode);
                this.status = "paid";
                break;

            default:
                throw new IllegalArgumentException("Unknown payment type: " + paymentType);
        }
    }

    public String getStatus() {
        return status;
    }
}
