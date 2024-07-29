          <li class="breadcrumb-item"><a href="#">Editing Item Category - <%= @item_category.name %></a></li>
        </ol>
      </div>
    </div>
  </div><!-- /.container-fluid -->
</section>
<!-- Main content -->
<%= render partial: "form", object: @item_category %>

