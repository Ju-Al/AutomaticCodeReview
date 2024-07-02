# frozen_string_literal: true

class Order
  include Mongoid::Document
  field :status, type: Mongoid::StringifiedSymbol
end