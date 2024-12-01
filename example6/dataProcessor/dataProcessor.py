class DataProcessor:
    def process_sales_data(self, sales_data):
        # Duplizierter Validierungscode
        if not sales_data:
            raise ValueError("Sales data cannot be empty")
        
        if not isinstance(sales_data, list):
            raise TypeError("Sales data must be a list")
        
        processed_sales = []
        for sale in sales_data:
            # Duplizierte Validierungslogik
            if 'amount' not in sale:
                raise ValueError(f"Invalid sale record: {sale}")
            
            if 'date' not in sale:
                raise ValueError(f"Invalid sale record: {sale}")
            
            # Komplexe Verarbeitungslogik
            processed_sale = {
                'amount': float(sale['amount']),
                'date': self.normalize_date(sale['date']),
                'is_valid': self.validate_sale(sale)
            }
            processed_sales.append(processed_sale)
        
        return processed_sales
    
    def process_purchase_data(self, purchase_data):
        # Fast identischer Code wie bei Sales Data
        if not purchase_data:
            raise ValueError("Purchase data cannot be empty")
        
        if not isinstance(purchase_data, list):
            raise TypeError("Purchase data must be a list")
        
        processed_purchases = []
        for purchase in purchase_data:
            # Fast gleiche Validierungslogik
            if 'amount' not in purchase:
                raise ValueError(f"Invalid purchase record: {purchase}")
            
            if 'date' not in purchase:
                raise ValueError(f"Invalid purchase record: {purchase}")
            
            # Ähnliche Verarbeitungslogik
            processed_purchase = {
                'amount': float(purchase['amount']),
                'date': self.normalize_date(purchase['date']),
                'is_valid': self.validate_purchase(purchase)
            }
            processed_purchases.append(processed_purchase)
        
        return processed_purchases
    
    def normalize_date(self, date):
        # Datumsformatierungslogik
        try:
            # Verschiedene Datumsformatierungen normalisieren
            return str(date)
        except Exception as e:
            raise ValueError(f"Invalid date format: {date}")
    
    def validate_sale(self, sale):
        # Verkaufsvalidierung
        return (
            sale.get('amount', 0) > 0 and 
            sale.get('date') is not None
        )
    
    def validate_purchase(self, purchase):
        # Kaufvalidierung
        return (
            purchase.get('amount', 0) > 0 and 
            purchase.get('date') is not None
        )

def main():
    processor = DataProcessor()
    
    sales_data = [
        {'amount': 100, 'date': '2023-01-01'},
        {'amount': 200, 'date': '2023-02-01'}
    ]
    
    purchase_data = [
        {'amount': 150, 'date': '2023-03-01'},
        {'amount': 250, 'date': '2023-04-01'}
    ]
    
    processed_sales = processor.process_sales_data(sales_data)
    processed_purchases = processor.process_purchase_data(purchase_data)
    
    print("Processed Sales:", processed_sales)
    print("Processed Purchases:", processed_purchases)

if __name__ == "__main__":
    main()