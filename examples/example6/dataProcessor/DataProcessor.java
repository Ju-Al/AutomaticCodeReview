package example6.dataProcessor;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Predicate;

class DataProcessor {
    public List<Map<String, Object>> processSalesData(List<Map<String, Object>> salesData) {
        // Duplizierter Validierungscode
        if (salesData == null || salesData.isEmpty()) {
            throw new IllegalArgumentException("Sales data cannot be empty");
        }

        List<Map<String, Object>> processedSales = new ArrayList<>();
        for (Map<String, Object> sale : salesData) {
            // Duplizierte Validierungslogik
            if (!sale.containsKey("amount")) {
                throw new IllegalArgumentException("Invalid sale record: " + sale);
            }

            if (!sale.containsKey("date")) {
                throw new IllegalArgumentException("Invalid sale record: " + sale);
            }

            // Komplexe Verarbeitungslogik
            Map<String, Object> processedSale = new HashMap<>();
            processedSale.put("amount", Double.parseDouble(sale.get("amount").toString()));
            processedSale.put("date", normalizeDate(sale.get("date")));
            processedSale.put("is_valid", validateSale(sale));

            processedSales.add(processedSale);
        }

        return processedSales;
    }

    public List<Map<String, Object>> processPurchaseData(List<Map<String, Object>> purchaseData) {
        // Fast identischer Code wie bei Sales Data
        if (purchaseData == null || purchaseData.isEmpty()) {
            throw new IllegalArgumentException("Purchase data cannot be empty");
        }

        List<Map<String, Object>> processedPurchases = new ArrayList<>();
        for (Map<String, Object> purchase : purchaseData) {
            // Fast gleiche Validierungslogik
            if (!purchase.containsKey("amount")) {
                throw new IllegalArgumentException("Invalid purchase record: " + purchase);
            }

            if (!purchase.containsKey("date")) {
                throw new IllegalArgumentException("Invalid purchase record: " + purchase);
            }

            // Ähnliche Verarbeitungslogik
            Map<String, Object> processedPurchase = new HashMap<>();
            processedPurchase.put("amount", Double.parseDouble(purchase.get("amount").toString()));
            processedPurchase.put("date", normalizeDate(purchase.get("date")));
            processedPurchase.put("is_valid", validatePurchase(purchase));

            processedPurchases.add(processedPurchase);
        }

        return processedPurchases;
    }

    private String normalizeDate(Object date) {
        // Datumsformatierungslogik
        try {
            return date.toString();
        } catch (Exception e) {
            throw new IllegalArgumentException("Invalid date format: " + date);
        }
    }

    private boolean validateSale(Map<String, Object> sale) {
        // Verkaufsvalidierung
        return sale.containsKey("amount") &&
                sale.containsKey("date") &&
                Double.parseDouble(sale.get("amount").toString()) > 0;
    }

    private boolean validatePurchase(Map<String, Object> purchase) {
        // Kaufvalidierung
        return purchase.containsKey("amount") &&
                purchase.containsKey("date") &&
                Double.parseDouble(purchase.get("amount").toString()) > 0;
    }

    public static void main(String[] args) {
        DataProcessor processor = new DataProcessor();

        List<Map<String, Object>> salesData = new ArrayList<>();
        Map<String, Object> sale1 = new HashMap<>();
        sale1.put("amount", 100);
        sale1.put("date", "2023-01-01");
        salesData.add(sale1);

        List<Map<String, Object>> purchaseData = new ArrayList<>();
        Map<String, Object> purchase1 = new HashMap<>();
        purchase1.put("amount", 150);
        purchase1.put("date", "2023-03-01");
        purchaseData.add(purchase1);

        List<Map<String, Object>> processedSales = processor.processSalesData(salesData);
        List<Map<String, Object>> processedPurchases = processor.processPurchaseData(purchaseData);

        System.out.println("Processed Sales: " + processedSales);
        System.out.println("Processed Purchases: " + processedPurchases);
    }
}
