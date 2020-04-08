import React, { Component } from 'react';
import { Layout, Menu, Icon, Breadcrumb } from 'antd';
import { slide as SlideMenu } from 'react-burger-menu';
import { NavLink } from 'react-router-dom';
import request from 'superagent';
import './layout.scss';

const { Header, Content, Footer } = Layout;
const { SubMenu } = Menu;

class DefaultMenu extends Component {
  render() {
    return <div>Empty Menu</div>;
  }
}

export default class CoreLayout extends Component {
  constructor(props) {
    super(props)
    const { route, match } = props;
    this.state = {
      selectedKeys: [route.name],
      name: route.name,
      namespaces: [],
      selectNamespace: match.params.namespace || 'default',
      slideMenuOpen: false,
      slideMenuComponent: <DefaultMenu />,
    }
    this.getNamespace();
  }

  componentDidMount() {
    const { route } = this.props
    this.setState({ selectedKeys: [route.name] });
  }

  getNamespace() {
    request
      .get('/api/cluster/namespaces')
      .then((res) => {
        if (res.body) {
          const namespaces = res.body.map((item) => {
            return item.metadata.name
          })
          this.setState({ namespaces })
        }
      })
  }

  onOpenSideMenuClick = ({ content: slideMenuComponent }) => {
    this.setState({ slideMenuOpen: true, slideMenuComponent })
  }

  render() {
    const { component: Component, ...props } = this.props;
    const MenuComponent = this.state.slideMenuComponent;
    return (
      <Layout>
        <Header style={{ position: 'fixed', zIndex: 2, width: '100%' }}>
          <div className="logo" />
          <Menu
            theme="dark"
            mode="horizontal"
            selectedKeys={this.state.selectedKeys}
            style={{ lineHeight: '64px' }}
          >
            <SubMenu
              title={
                <span className="submenu-title-wrapper">
                  <Icon type="deployment-unit" />
                  {this.state.selectNamespace}
                </span>
              }
            >
              {
                this.state.namespaces
                  .map((item) => (<Menu.Item key={item}><a href={`/${item}/dashboard`}>{item}</a></Menu.Item>))
              }
            </SubMenu>
            <Menu.Item key="Visual">
              <NavLink to={`/${this.state.selectNamespace}/visual`}>Visual</NavLink>
            </Menu.Item>
            <SubMenu
              title={<span><Icon type="question-circle" />help</span>}
              className="pull-right"
            >
              <Menu.Item key="feedback"><a href="https://github.com/rdc-incubator/kt-docs/issues/new" target="_blank" rel="noopener noreferrer">Feedback</a></Menu.Item>
              <Menu.Item key="doc"><a href="https://rdc-incubator.github.io/kt-docs/#/" target="_blank" rel="noopener noreferrer">Documents</a></Menu.Item>
              <Menu.Item key="downloads"><a href="https://rdc-incubator.github.io/kt-docs/#/downloads" target="_blank" rel="noopener noreferrer">Cli Download</a></Menu.Item>
            </SubMenu>
          </Menu>
        </Header>
        <SlideMenu
          right
          width={'30%'}
          isOpen={this.state.slideMenuOpen}
          onStateChange={(state) => {
            this.setState({ slideMenuOpen: state.isOpen })
          }}
        >
          {MenuComponent}
        </SlideMenu>
        <Content style={{ padding: '0 50px', marginTop: 64 }}>
          <Breadcrumb style={{ margin: '16px 0' }}>
            <Breadcrumb.Item>{this.state.name}</Breadcrumb.Item>
          </Breadcrumb>
          <div style={{ minHeight: 380 }}>
            <Component {...props} onOpenSideMenuClick={this.onOpenSideMenuClick} />
          </div>
        </Content>
        <Footer style={{ textAlign: 'center' }}>KT Connect@2019</Footer>

      </Layout>
    );
  }

}