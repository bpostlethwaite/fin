const config = require('./config')();

var chromedriver = require('chromedriver');
module.exports = {

  finconfig: config,

  before: function(done) {

    chromedriver.start();

    done();
  },

  after: function(done) {
    chromedriver.stop();

    done();
  }
};
