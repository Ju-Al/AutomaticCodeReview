                
                var prettyDropdown = s.Driver.FindElement(By.Id("prettydropdown-DefaultLang"));
                prettyDropdown.Click();
                await Task.Delay(200);
                prettyDropdown.FindElement(By.CssSelector("[data-value=\"da-DK\"]")).Click();
                await Task.Delay(1000);
                Assert.NotEqual(payWithTextEnglish, s.Driver.FindElement(By.Id("pay-with-text")).Text);
                s.Driver.Navigate().GoToUrl(s.Driver.Url + "?lang=da-DK");
                
                Assert.NotEqual(payWithTextEnglish, s.Driver.FindElement(By.Id("pay-with-text")).Text);
                
                s.Driver.Quit();
            }
        }
        
        [Fact]
        public void CanUsePaymentMethodDropdown()
        {
            using (var s = SeleniumTester.Create())
            {
                s.Start();
                s.RegisterNewUser();
                var store = s.CreateNewStore();
                s.AddDerivationScheme("BTC");
                
                //check that there is no dropdown since only one payment method is set
                var invoiceId = s.CreateInvoice(store.storeName, 10, "USD", "a@g.com");
                s.GoToInvoiceCheckout(invoiceId);
                s.Driver.FindElement(By.ClassName("payment__currencies_noborder"));
                s.GoToHome();
                s.GoToStore(store.storeId);
                s.AddDerivationScheme("LTC");
                s.AddLightningNode("BTC",LightningConnectionType.CLightning);
                //there should be three now
                invoiceId = s.CreateInvoice(store.storeName, 10, "USD", "a@g.com");
                s.GoToInvoiceCheckout(invoiceId);
                var currencyDropdownButton =  s.Driver.FindElement(By.ClassName("payment__currencies"));
                Assert.Contains("BTC", currencyDropdownButton.Text);
                currencyDropdownButton.Click();
                
                var elements = s.Driver.FindElement(By.ClassName("vex-content"))
                    .FindElements(By.ClassName("vexmenuitem"));
                Assert.Equal(3, elements.Count);
                elements.Single(element => element.Text.Contains("LTC")).Click();
                currencyDropdownButton =  s.Driver.FindElement(By.ClassName("payment__currencies"));
                Assert.Contains("LTC", currencyDropdownButton.Text);
                
                elements = s.Driver.FindElement(By.ClassName("vex-content"))
                    .FindElements(By.ClassName("vexmenuitem"));
                
                elements.Single(element => element.Text.Contains("Lightning")).Click();
                
                Assert.Contains("Lightning", currencyDropdownButton.Text);
                
                s.Driver.Quit();
            }
        }
        
        [Fact]
        public void CanUseLightningSatsFeature()
        {
            //uncomment after https://github.com/btcpayserver/btcpayserver/pull/1014
//            using (var s = SeleniumTester.Create())
//            {
//                s.Start();
//                s.RegisterNewUser();
//                var store = s.CreateNewStore();
//                s.AddInternalLightningNode("BTC");
//                s.GoToStore(store.storeId, StoreNavPages.Checkout);
//                s.SetCheckbox(s, "LightningAmountInSatoshi", true);
//                s.Driver.FindElement(By.Name("command")).Click();
//                var invoiceId = s.CreateInvoice(store.storeName, 10, "USD", "a@g.com");
//                s.GoToInvoiceCheckout(invoiceId);
//                Assert.Contains("Sats", s.Driver.FindElement(By.ClassName("payment__currencies_noborder")).Text);
//                
//            }
        }
    }
}
