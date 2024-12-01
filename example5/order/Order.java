package example5.order;

import java.util.ArrayList;
import java.util.List;

public class Order {
    private List<String> items;
    private List<Integer> quantities;
    private List<Double> prices;
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

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }
}
