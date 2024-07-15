import streamlit as st
import yaml
def load_config():
    try:
        with open('config.yaml', 'r') as file:
            return yaml.safe_load(file)
    except FileNotFoundError:
        return {'listen_port': 8080, 'routes': {}}

def save_config(config):
    with open('config.yaml', 'w') as file:
        yaml.dump(config, file)

st.title('Reverse Proxy Configuration')

config = load_config()

# Listen Port
listen_port = st.number_input('Listen Port', min_value=1, max_value=65535, value=config.get('listen_port', 8080))

# Routes
st.subheader('Routes')
routes = config.get('routes', {})

# Display existing routes
for host, target in routes.items():
    col1, col2, col3 = st.columns([2, 2, 1])
    with col1:
        st.text(host)
    with col2:
        st.text(target)
    with col3:
        if st.button('Delete', key=f'delete_{host}'):
            del routes[host]

# Add new route
st.subheader('Add New Route')
new_host = st.text_input('Host (e.g., example.com)')
new_target = st.text_input('Target (e.g., http://localhost:9000)')
if st.button('Add Route'):
    if new_host and new_target:
        routes[new_host] = new_target
        st.success(f'Added route: {new_host} -> {new_target}')
    else:
        st.error('Both host and target are required')

# Save configuration
if st.button('Save Configuration'):
    config['listen_port'] = listen_port
    config['routes'] = routes
    save_config(config)
    st.success('Configuration saved successfully!')

# Display current configuration
st.subheader('Current Configuration')
st.code(yaml.dump(config), language='yaml')
