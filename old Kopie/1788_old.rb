<!doctype html>
        <h1 class="title">Get in touch to sign up</h1>
        <h2 class="sub_title">start the process by clicking below or sending us an email at <a href="mailto:info@diaper.app?Subject=Diaperbase%20Interest" target="_top" >info@diaper.app</a></h2>
        <a href="mailto:info@diaper.app?Subject=Diaperbase%20Interest" target="_top" class="btn_fancy">
          <div class="solid_layer"></div>
          <div class="text_layer">Get In Touch!</div>
        </a>
        <!--    End Button    -->
<html class="no-js" lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>DiaperBase</title>
    
    <!--    Stylesheet Files    -->
    <link rel="stylesheet" href="css/normalize.css" />
    <link rel="stylesheet" href="css/foundation.min.css" />
    <link rel="stylesheet" href="css/main.css" />


    <!--    Javascript files are placed before </body>    -->
  </head>
  
  <body>
    <!--  Start Hero Section  -->
    <section class="hero">
      <header>
        <div class="row">
          

          <nav class="top-bar" data-topbar role="navigation">
            
            <!--    Start Logo    -->
            <ul class="title-area">
              <li class="name">
                <a href="#" class="logo">
                  <h1>diaper<span class="tld">.app</span></h1>
                </a>
              </li>
                <span class="toggle-topbar menu-icon"><a href="#"><span>Menu</span></a></span>
              </li>
            </ul>  
            <!--    End Logo    -->

            <!--    Start Navigation Menu    -->
            <section class="top-bar-section" id="mean_nav">
              <ul class="right">
                <li><a href="#services">About</a></li>
                <li><a href="#testimonials">Testimonials</a></li>
                <li><a href="#contact">Contact</a></li>
                <li><%= link_to 'User Sign In', new_user_session_path(organization_id: nil) %></li>
              </ul>
            </section>
            <!--    End Navigation Menu    -->

          </nav>
        </div>
      </header>

      <!--    Start Hero Caption    -->
      <section class="caption">
        <div class="row">
          <h1 class="superbar" style="text-align: center;">
          <img id="MastheadLogo" src="/img/diaper-base-logo-full-white-text.svg" style="width: 70%; height: 70%; " alt="Diaperbase" title="Diaperbase" class="serv_icon"/>
          </h1>
          <h2 class="superbar">The easiest and most love-filled way to manage your diaper bank.</h2>
        </div>
      </section>
      <!--    End Hero Caption    -->

    </section>
    <!--  End Hero Section  -->



    <!--  Start Services Section  -->
    <section class="services" id="services">     

      <!--    Start Services Titles    -->
      <div class="row">        
        <h1 class="mean_title">Diaperbase removes the complexity from your day</h1>
        <h2 class="sub_title">and lets you spend time helping those who need it.</h2> 
      </div>
      <!--    End Services Titles    -->

      <!--    Start Services List    -->
      <div class="row services_list">
        <div class="small-12 medium-3 large-3 columns">
          <img src="/img/icon2.png" alt="" title="" height="200" width="200" class="serv_icon"/>
          <h2 class="title">Report Generation</h2>
          <p>Generate pdfs of partner distribution invoices for easy printing and sending.</p>
        </div>

        <div class="small-12 medium-3 large-3 columns">
          <img src="/img/icon1.png" alt="" title="" height="200" width="200" class="serv_icon"/>
          <h2 class="title">Barcode Scanning</h2>
          <p>Quickly add or remove inventory with the scan of a barcode. The adjusting and transfering of
          inventory is also supported as well as creating your own barcodes.</p>
        </div>

        <div class="small-12 medium-3 large-3 columns">
          <img src="/img/icon3.png" alt="" title="" height="200" width="200" class="serv_icon"/>
          <h2 class="title">Inventory Management</h2>
          <p>Easily keep track of the items in your inventory regardless of their storage location.</p>
        </div>

        <div class="small-12 medium-3 large-3 columns">
          <img src="/img/icon4.png" alt="" title="" height="200" width="200" class="serv_icon"/>
          <h2 class="title">Dashboard Visualization</h2>
          <p>Quickly view your inventory, donations, purchases and distributions on a single page.</p>
        </div>
      </div>
      <!--    End Services List    -->

      <!--    Start Button    -->
      <div class="btn_holder">
        <a href="mailto:info@diaper.app?Subject=Diaperbase%20Interest" target="_top" class="btn_fancy">
          <div class="solid_layer"></div>
          <div class="text_layer">Get In Touch!</div>
        </a>
      </div>
      <!--    End Button    -->

    </section>
    <!--  End Services Section  -->



    <!--  Start Quote Section  -->
    <section class="quote">
        <blockquote>
          <p>No one is useless in this world who lightens the burdens of another.</p>
          <hr/>
          <span class="author">Charles Dickens</span>
        </blockquote>
    </section>
    <!--  End Quote Section  -->

    <!--  Start Testimonials Section  -->
    <section class="testimonials" id="testimonials">
      <div class="row">
        <div class="slider_container">
          <div id="carousel">
            
            <div class="tesimonial">
              <img src="/img/diaper-base-logo-icon.svg" title="Diaperbase Logo" alt="Diaperbase Logo">
              <!-- <span class="name">Mashable</span> -->
              <p>Finally, an inventory management system designed specifically for the unique needs of diaper banks!</p>
              <span class="author">Rachel Alston, PDX Diaper Bank</span>
            </div>

            <div class="tesimonial">
              <img src="/img/diaper-base-logo-icon.svg" title="Diaperbase Logo" alt="Diaperbase Logo">
              <!-- <span class="name">Mashable_2</span> -->
              <p>DiaperBase has taken our diaper distribution to the next level. We no longer have to worry about how to track our incoming and outgoing inventory, and having all our products and data tracked in different ways in different documents and systems. DiaperBase has saved us time, money, energy and many, many headaches. The system is extremely intuitive and user-friendly and we are so grateful to have it.</p>
              <span class="author">Megan, Sweet Cheeks Diaper Bank</span>
            </div>

          </div>
        </div>

        <!--    Start Testimonials Pagination    -->
        <nav class="pagination">
          <ul>
            <li><a href="#">1</a></li>
            <li><a href="#" class="selected">2</a></li>
          </ul>
        </nav>
        <!--    End Testimonials Pagination    -->

      </div>
    </section>
    <!--  End Testimonials Section  -->



    <!--  Start Call To Action Section  -->
    <section class="cta" id="contact">
      <div class="row">

        <!--    Start CTA Titles    -->
        <h1 class="title">Want to use Diaperbase? </h1>
        <!--    End CTA Titles    -->

        <!--    Start Button    -->
        <%= button_to "Click Here To Request A Demo", new_account_request_path,  method: :get%>
        <!--
          End
          Button
        -->
        <br>
        <p>
          <strong>After the demo, you will be able to request your own account.</strong>
          <br>
          Have questions? Email us at <a href="mailto:info@diaper.app?Subject=Diaperbase%20Interest" target="_top" >info@diaper.app</a>
        </p>
      </div>
    </section>
    <!--  End Call To Action Section  -->


    <!--  Start Footer Section  -->
    <footer>
      <div class="row">
        
        <!--    Start Copyrights    -->
        <div class="small-12 medium-4 large-4 columns">
          <div class="copyrights">
            <a class="logo" href="#">
              <h1>diaper<span class="tld">.app</span></h1>
            </a>
          </div>
        </div>
        <!--    End Copyrights    -->


        <div class="small-12 medium-8 large-8 columns">
          <div class="contact_details right">
            <nav class="social">
              <ul class="no-bullet">

              </ul>
            </nav>

            <div class="contact">
                <p>Diaperbase was lovingly</p>
                <p>swaddled by:</p>

              <p><a href="http://www.pdxdiaperbank.org/">PDX Diaper Bank</a> &</p>
              <p><a href="http://rubyforgood.org/">Ruby for Good</a></p>
            </div>
          </div>
        </div>

      </div>
    </footer>
    <!--  End Footer Section  -->

    <!--    Javascript Files    -->
    <script type="text/javascript" src="js/jquery.js"></script>
    <script type="text/javascript" src="js/touchSwipe.min.js"></script>
    <script type="text/javascript" src="js/easing.js"></script>
    <script type="text/javascript" src="js/foundation.min.js"></script>
    <script type="text/javascript" src="js/foundation/foundation.topbar.js"></script>
    <script type="text/javascript" src="js/carouFredSel.js"></script>
    <script type="text/javascript" src="js/scrollTo.js"></script>
    <script type="text/javascript" src="js/map.js"></script>
    <script type="text/javascript" src="js/main.js"></script>

  </body>
</html>
