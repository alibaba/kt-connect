import React, { Component } from 'react';
import Point from '../../../components/Point'
import Menus from './Menus';
import './graph.scss'

class Graph extends Component {
  render() {
    const { item, x } = this.props;
    return (
      <div>
        <div className="root-style">
          <Point
            title={item.metadata.name}
            className={item.proxyToLocal ? 'hightlight' : 'normal'}
            icon="cluster"
          />
        </div>
        <div>
          {
            item.subsets.map((subset, y) => {
              return subset.addresses ? (
                <div className="row-style" key={`subset-${x}-${y}`}>{subset.addresses.map((address, z) => {
                  return (
                    <Point
                      key={`address-${x}-${y}-${z}`}
                      title={address.ip}
                      className={address.proxyStatus ? 'hightlight' : 'normal'}
                      icon={address.proxyStatus ? 'cloud-download' : ''}
                      onClick={() => {
                        if (!address.proxyStatus) return;
                        const type = address.proxyStatus ? 'component' : 'pod';
                        const Menu = Menus({ type, component: address.component ? address.component[0] : null });
                        this.props.onOpenSideMenuClick({
                          content: Menu
                        });
                      }}
                    />
                  )
                }
                )}
                </div>
              ) : null
            })
          }
        </div>
      </div>
    );
  }
}

export default Graph;
