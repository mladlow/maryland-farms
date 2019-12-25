#!/usr/bin/env python
import os
os.environ['AWS_SHARED_CREDENTIALS_FILE'] = './.aws/credentials'

import csv
import json
import time
import os
import urllib2
import boto3

FILENAME = './data/LicensedStables2015.csv'
GOOGLE_API_KEY = './data/GoogleApiKey.json'

DYNAMO_TABLE_NAME = 'maryland_farms';

COUNTIES = {
        'AL': 'Allegany County',
        'AA': 'Anne Arundel County',
        'BL': 'Baltimore County',
        'BA': 'Baltimore County',
        'BC': 'Baltimore City',
        'CV': 'Calvert County',
        'CL': 'Caroline County',
        'CR': 'Carroll County',
        'CC': 'Cecil County',
        'CE': 'Cecil County',
        'CH': 'Charles County',
        'DR': 'Dorchester County',
        'FR': 'Frederick County',
        'GR': 'Garrett County',
        'HF': 'Harford County',
        'HR': 'Harford County',
        'HW': 'Howard County',
        'KN': 'Kent County',
        'KT': 'Kent County',
        'MG': 'Montgomery County',
        'PG': 'Prince George\'s County',
        'QA': 'Queen Anne\'s County',
        'SM': 'St. Mary\'s County',
        'SS': 'Somerset County',
        'SO': 'Somerset County',
        'TA': 'Talbot County',
        'TB': 'Talbot County',
        'WA': 'Washington County',
        'WH': 'Washington County',
        'WC': 'Wicomico County',
        'WI': 'Wicomico County',
        'WO': 'Worcester County',
        'WR': 'Worcester County',
        }

# This script parses a CSV, tries to geocode the address, and writes to dynamodb.
def read_csv():
    stable_list = []
    with open(FILENAME, 'rU') as csv_file:
        reader = csv.reader(csv_file, strict=True, delimiter='\t')
        # Skip header
        next(reader, None)

        for row in reader:
            if len(row) != 8 or row[2] not in COUNTIES:
                print 'Invalid row', row
                raise
            county_pretty = COUNTIES[row[2]]
            # print row[0:2] + [county_pretty] + row[3:]
            stable_list.append(row[0:2] + [county_pretty] + row[3:])
    return stable_list

def extract_api_key():
    key = ''
    with open(GOOGLE_API_KEY, 'r') as json_file:
        key = json.load(json_file)['GOOGLE_API_KEY']
    return key

def initialize_dynamo():
    test_session = boto3.Session()
    creds = test_session.get_credentials()
    current_creds = creds.get_frozen_credentials()
    print current_creds.access_key
    dynamo_resource = boto3.resource('dynamodb',
            region_name='us-east-1')
    return dynamo_resource.Table(DYNAMO_TABLE_NAME)

def geocode_address(api_key, address_array):
    url = 'https://maps.googleapis.com/maps/api/geocode/json'
    address_param = 'address=' + ','.join([x.replace(' ', '+') for x in address_array])
    key_param = 'key=' + api_key
    load_from = url + '?' + address_param + '&' + key_param

    geocoded = json.load(urllib2.urlopen(load_from))
    return geocoded

def add_to_dynamo(stable, geocoded_data, dyanmo_table):
    # Stop if no records
    if len(geocoded_data['results']) < 1:
        print 'Warning - bad results for ' + stable[1]
        print stable
        print json.dumps(geocoded_data, indent=4)
        return
    # Just use the first if more than 1
    if len(geocoded_data['results']) > 1:
        print 'Warning - multiple results for ' + stable[1]
        print stable
        print json.dumps(geocoded_data, indent=4)
    out_data = {
            'id': stable[0],
            'title': stable[1],
            'position': geocoded_data['results'][0]['geometry']['location'],
            'address': geocoded_data['results'][0]['formatted_address'],
            'phone': stable[-1],
            }
    #keep_on = raw_input('Write to file? ')
    keep_on = 'y'
    if (keep_on == 'y'):
        with open('./data/json/' + stable[0] + '.json', 'w') as out_file:
            json.dump(out_data, out_file)

def process_data():
    stable_list = read_csv()
    api_key = extract_api_key()
    dynamo_table = ''
    #dynamo_table = initialize_dynamo()
    #print dynamo_table.creation_date_time
    for stable in stable_list:
        if os.path.isfile('./data/json/' + stable[0] + '.json'):
            continue
        print stable
        #keep_on = raw_input('Geocode? ')
        keep_on = 'y'
        if (keep_on == 'y'):
            geocoded = geocode_address(api_key, stable[3:7])
            #print json.dumps(geocoded, indent=4)
            add_to_dynamo(stable, geocoded, dynamo_table)
            time.sleep(5)

if __name__ == '__main__':
    process_data()
    #geocoded = {u'status': u'OK', u'results': [{u'geometry': {u'location': {u'lat': 39.444818, u'lng': -76.979773}, u'viewport': {u'northeast': {u'lat': 39.4461669802915, u'lng': -76.9784240197085}, u'southwest': {u'lat': 39.4434690197085, u'lng': -76.9811219802915}}, u'location_type': u'ROOFTOP'}, u'address_components': [{u'long_name': u'4785', u'types': [u'street_number'], u'short_name': u'4785'}, {u'long_name': u'Bartholow Road', u'types': [u'route'], u'short_name': u'Bartholow Rd'}, {u'long_name': u'Eldersburg', u'types': [u'locality', u'political'], u'short_name': u'Eldersburg'}, {u'long_name': u'14, Berrett', u'types': [u'administrative_area_level_3', u'political'], u'short_name': u'14, Berrett'}, {u'long_name': u'Carroll County', u'types': [u'administrative_area_level_2', u'political'], u'short_name': u'Carroll County'}, {u'long_name': u'Maryland', u'types': [u'administrative_area_level_1', u'political'], u'short_name': u'MD'}, {u'long_name': u'United States', u'types': [u'country', u'political'], u'short_name': u'US'}, {u'long_name': u'21784', u'types': [u'postal_code'], u'short_name': u'21784'}, {u'long_name': u'9204', u'types': [u'postal_code_suffix'], u'short_name': u'9204'}], u'place_id': u'ChIJJQ9jCfE6yIkRw91NVhv1GiQ', u'formatted_address': u'4785 Bartholow Rd, Eldersburg, MD 21784, USA', u'types': [u'street_address']}]}
