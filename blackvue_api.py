import requests


class BlackVueAPI:
    URL = 'https://api.blackvuecloud.com/BCS'

    def __init__(self, api_key):
        self.user_token = api_key

    # register_cam registers a camera with the BlackVue Cloud API under a specified user account 
    # and returns a json response containing api resonse
    #   Params:
    #       Email email:            email associated with camera
    #       int device_serial:      serial number of camera
    #       int cloud_code:         cloud registration code for camera
    #       string dev_name:        string name of device
    #       boolean agree_gps:      denotes if gps data is captured
    #   Returns:
    #       requests.Response:      response object containing server's response
    #                               to an HTTP request
    def register_cam(self, email, device_serial, cloud_code, dev_name, agree_gps):
        endpoint = '/device_register.pop'
        payload = {'email': email, 'user_token': self.user_token, 'psn': device_serial,
                   'cldn': cloud_code, 'dev_name': dev_name, 'agree_gps': agree_gps}
        r = requests.get(self.URL + endpoint, params=payload)
        return r

    def request_vod_list_json(self, email, device_serial):
        endpoint = '/proc/vod_list'
        payload = {'email': email, 'user_token': self.user_token,
                   'psn': device_serial}
        r = requests.get(self.URL + endpoint, params=payload)
        return r

    def request_vod_mp4_from_cam(self, email, device_serial, vod_token):
        endpoint = '/proc/vod_file'
        payload = {'email': email, 'user_token': self.user_token, 'psn': device_serial, , vod_token = vod_token}
        r = requests.get(self.URL + endpoint, params=payload)
        return r

    def request_cam_gps_data(self, self, email, device_serial):
        endpoint = '///gps_zone.php'
        payload = {'email': email, 'user_token': self.user_token,
                   'psn': device_serial}
        r = requests.get(self.URL + endpoint, params=payload)
        return r

    def add_user(self):  # TODO: implement once logic behind user vs device is understood
        pass

    def user_login(self):  # TODO: implement once logic behind user vs device is understood
        pass

    def user_logout(self):  # TODO: implement once logic behind user vs device is understood
        pass


# TODO: experiment with calls on cameras to ensure functions are working
# TODO: complete descriptions for every argument