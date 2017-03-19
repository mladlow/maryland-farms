#!/usr/bin/env python
import os
os.environ['AWS_SHARED_CREDENTIALS_FILE'] = './.aws/credentials'

import csv
import json
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
    csv_file = open(FILENAME, 'rU')
    reader = csv.reader(csv_file, strict=True)
    # Skip header
    next(reader, None)
    
    stable_list = []
    for row in reader:
        if len(row) != 8 or row[2] not in COUNTIES:
            print 'Invalid row', row
            raise
        county_pretty = COUNTIES[row[2]]
        # print row[0:2] + [county_pretty] + row[3:]
        stable_list.append(row[0:2] + [county_pretty] + row[3:])
    return stable_list

def extract_api_key():
    json_file = open(GOOGLE_API_KEY, 'r')
    return json.load(json_file)['GOOGLE_API_KEY']

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
    print stable
    print geocoded_data
    # TODO
    item = {
            'stable_name': stable[1],
            }

def process_data():
    stable_list = read_csv()
    api_key = extract_api_key()
    dynamo_table = initialize_dynamo()
    print dynamo_table.creation_date_time
    # geocoded = geocode_address(api_key, stable_list[0][3:7])
    geocoded = {u'status': u'OK', u'results': [{u'geometry': {u'location': {u'lat': 39.444818, u'lng': -76.979773}, u'viewport': {u'northeast': {u'lat': 39.4461669802915, u'lng': -76.9784240197085}, u'southwest': {u'lat': 39.4434690197085, u'lng': -76.9811219802915}}, u'location_type': u'ROOFTOP'}, u'address_components': [{u'long_name': u'4785', u'types': [u'street_number'], u'short_name': u'4785'}, {u'long_name': u'Bartholow Road', u'types': [u'route'], u'short_name': u'Bartholow Rd'}, {u'long_name': u'Eldersburg', u'types': [u'locality', u'political'], u'short_name': u'Eldersburg'}, {u'long_name': u'14, Berrett', u'types': [u'administrative_area_level_3', u'political'], u'short_name': u'14, Berrett'}, {u'long_name': u'Carroll County', u'types': [u'administrative_area_level_2', u'political'], u'short_name': u'Carroll County'}, {u'long_name': u'Maryland', u'types': [u'administrative_area_level_1', u'political'], u'short_name': u'MD'}, {u'long_name': u'United States', u'types': [u'country', u'political'], u'short_name': u'US'}, {u'long_name': u'21784', u'types': [u'postal_code'], u'short_name': u'21784'}, {u'long_name': u'9204', u'types': [u'postal_code_suffix'], u'short_name': u'9204'}], u'place_id': u'ChIJJQ9jCfE6yIkRw91NVhv1GiQ', u'formatted_address': u'4785 Bartholow Rd, Eldersburg, MD 21784, USA', u'types': [u'street_address']}]}
    add_to_dynamo(stable_list[0], geocoded, dynamo_table)

if __name__ == '__main__':
    process_data()
