import React, { Component } from 'react';
import PropTypes from 'prop-types';

import User from './user';

class UserList extends Component {
  render() {
    return (
      <ul>
        {this.props.users.map(user => (
          <User key={user.id} user={user} />
        ))}
      </ul>
    );
  }
}

UserList.propTypes = {
  users: PropTypes.array.isRequired
};

export default UserList;
