
     subject.post_hangout_notification(hangout)

     assert_requested(:post, 'https://agile-bot.herokuapp.com/hubot/hangouts-notify', times: 1) do |req|
       expect(req.body).to eq 'title=MockEvent&link=mock_url&type=PairProgramming'
     end
   end
end
