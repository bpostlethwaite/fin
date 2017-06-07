const MAX_WAIT = 4000;
const pages = {
  personal: {
    url: 'http://www.rbcroyalbank.com/personal.html',
    selectors: {
      signinBtn: '#layout-column-left button'
    }
  },
  portal: {
    selectors: {
      usernameInput: '#mainContent .ccUsername',
      passwordInput: '#mainContent input[type=password]',
      submitBtn: '#mainContent button[type=submit]'
    }
  },
  accounts: {
    selectors: {
      ccAccount: '#creditCards tbody tr:nth-child(2) th a'
    }
  },
  account: {
    selectors: {
      downloadLink: '.ccadPostedTable .tableHeader .tableHeaderMid a'
    }
  },
  download: {
    selectors: {
      csvRadio: 'input[value=EXCEL]',
      accountOption: '#accountInfo option:nth-child(3n)',
      allTransactions: '#transactionDropDown option:nth-child(2n)',
      downloadBtn: '#id_btn_continue'
    }
  }
};

module.exports = {
  'Open signin page': function(browser) {
    const page = pages.personal;
    browser
      .url(page.url)
      .waitForElementVisible(page.selectors.signinBtn, MAX_WAIT)
      .click(page.selectors.signinBtn);
  },
  'Sign in to accounts': function(browser) {
    const page = pages.portal;
    const config = browser.globals.finconfig;
    const username = config.banks.rbc.username;
    const password = config.banks.rbc.password;

    browser
      .waitForElementVisible(page.selectors.usernameInput, MAX_WAIT)
      .setValue(page.selectors.usernameInput,
                [username, browser.Keys.TAB])
      .setValue(page.selectors.passwordInput,
                [password, browser.Keys.TAB])
      .click(page.selectors.submitBtn);
  },
  'Open account': function(browser) {
    const page = pages.accounts;
    browser
      .waitForElementVisible(page.selectors.ccAccount, MAX_WAIT)
      .click(page.selectors.ccAccount);
  },
  'Open download page': function(browser) {
    const page = pages.account;
    browser
      .waitForElementVisible(page.selectors.downloadLink, MAX_WAIT)
      .click(page.selectors.downloadLink);
  },
  'download csv': function(browser) {
    const page = pages.download;
    browser
      .waitForElementVisible(page.selectors.csvRadio, MAX_WAIT)
      .click(page.selectors.csvRadio)
      .click(page.selectors.accountOption)
      .click(page.selectors.allTransactions)
      .click(page.selectors.downloadBtn);
  },
  'Complete': function(browser) {
    browser
      .pause(MAX_WAIT)
      .end();
  }
};
