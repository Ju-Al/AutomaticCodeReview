module RSpec
    RSpec.describe DidYouMean do
      describe '#call' do
        if defined?(::DidYouMean::SpellChecker)
          context "when `DidYouMean::SpellChecker` is available", skip: !defined?(::DidYouMean::SpellChecker) do
            context 'Success' do
              let(:name) { './spec/rspec/core/did_you_mean_spec.rb' }
              it 'returns a useful suggestion' do
                expect(DidYouMean.new(name[0..-2]).call).to include name
              end
              context 'numerous possibilities' do
                it 'returns a small number of suggestions' do
                  name = './spec/rspec/core/drb_spec.r'
                  suggestions = DidYouMean.new(name).call
                  expect(suggestions.split("\n").size).to eq 2
                end
              end
            end
            context 'No suitable suggestions' do
              it 'returns empty string' do
                name = './' + 'x' * 50
                expect(DidYouMean.new(name).call).to eq ''
              end
            end
          end
          context "when `DidYouMean::SpellChecker` is not available" do
            before do
              hide_const("::DidYouMean::SpellChecker") if defined?(::DidYouMean::SpellChecker)
            end
            describe 'Success' do
              let(:name) { './spec/rspec/core/did_you_mean_spec.rb' }
              it 'returns a hint' do
                expect(DidYouMean.new(name[0..-2]).call).to include 'Hint:'
              end
            end
          end
        end
      end
    end
  end
end
