      }, 300);
    };
  },
  
  afterConstruct: function(self) {
    apos.notify = self.trigger;
    
    self.createContainer();
  }
});
