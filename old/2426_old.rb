<section class="content-header">
  <div class="container-fluid">
    <div class="row mb-2">
      <div class="col-sm-6">
        <% content_for :title, "Edit - Item Category - #{current_organization.name}" %>
        <h1>
          Update record for <small><%= @item_category.name %></small>
        </h1>
      </div>
      <div class="col-sm-6">
        <ol class="breadcrumb float-sm-right">
          <li class="breadcrumb-item"><%= link_to(dashboard_path) do %>
              <i class="fa fa-dashboard"></i> Home
            <% end %>
          </li>
          <li class="breadcrumb-item"><%= link_to "All Item Types", (item_categories_path) %></li>
          <li class="breadcrumb-item"><a href="#">Editing Item Category - <%= @item_category.name %></a></li>
        </ol>
      </div>
    </div>
  </div><!-- /.container-fluid -->
</section>
<!-- Main content -->
<%= render partial: "form", object: @item_category %>

